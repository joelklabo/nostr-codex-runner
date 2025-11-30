package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// Runner wires transports, agent, and actions together.
type Runner struct {
	transports   []Transport
	transportMap map[string]Transport
	agent        Agent
	actions      map[string]Action
	actionSpecs  []ActionSpec
	logger       *slog.Logger

	reqTimeout    time.Duration
	actionTimeout time.Duration

	allowedActions map[string]struct{}
	allowedSenders map[string]struct{}

	auditStore AuditLogger
}

// AuditLogger records action executions.
type AuditLogger interface {
	AppendAudit(action, sender, outcome string, dur time.Duration) error
}

// RunnerOption configures a Runner.
type RunnerOption func(*Runner)

// WithRequestTimeout overrides the per-agent request timeout.
func WithRequestTimeout(d time.Duration) RunnerOption {
	return func(r *Runner) { r.reqTimeout = d }
}

// WithActionTimeout overrides the per-action timeout.
func WithActionTimeout(d time.Duration) RunnerOption {
	return func(r *Runner) { r.actionTimeout = d }
}

// WithAllowedActions sets a whitelist of action names; empty means allow all.
func WithAllowedActions(names []string) RunnerOption {
	set := make(map[string]struct{}, len(names))
	for _, n := range names {
		set[n] = struct{}{}
	}
	return func(r *Runner) { r.allowedActions = set }
}

// WithAllowedSenders sets allowed sender ids; empty means allow all.
func WithAllowedSenders(ids []string) RunnerOption {
	set := make(map[string]struct{}, len(ids))
	for _, n := range ids {
		set[strings.ToLower(n)] = struct{}{}
	}
	return func(r *Runner) { r.allowedSenders = set }
}

// WithAuditLogger wires an audit sink.
func WithAuditLogger(a AuditLogger) RunnerOption {
	return func(r *Runner) { r.auditStore = a }
}

// NewRunner constructs a Runner. If logger is nil, slog.Default is used.
func NewRunner(transports []Transport, agent Agent, actions []Action, logger *slog.Logger, opts ...RunnerOption) *Runner {
	if logger == nil {
		logger = slog.Default()
	}

	tmap := make(map[string]Transport, len(transports))
	for _, t := range transports {
		tmap[t.ID()] = t
	}

	amap := make(map[string]Action, len(actions))
	specs := make([]ActionSpec, 0, len(actions))
	for _, a := range actions {
		amap[a.Name()] = a
		specs = append(specs, ActionSpec{
			Name:         a.Name(),
			Capabilities: a.Capabilities(),
		})
	}

	r := &Runner{
		transports:    transports,
		transportMap:  tmap,
		agent:         agent,
		actions:       amap,
		actionSpecs:   specs,
		logger:        logger,
		reqTimeout:    15 * time.Minute,
		actionTimeout: 2 * time.Minute,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Start launches transports and processes inbound messages until ctx is done.
func (r *Runner) Start(ctx context.Context) error {
	inbound := make(chan InboundMessage, 128)
	var wg sync.WaitGroup
	errCh := make(chan error, len(r.transports))

	for _, t := range r.transports {
		wg.Add(1)
		go func(tr Transport) {
			defer wg.Done()
			if err := tr.Start(ctx, inbound); err != nil {
				errCh <- fmt.Errorf("transport %s: %w", tr.ID(), err)
			}
		}(t)
	}

	// Processor loop
	go func() {
		<-ctx.Done()
		close(inbound)
	}()

	for msg := range inbound {
		r.handleMessage(ctx, msg)
	}

	wg.Wait()

	select {
	case err := <-errCh:
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	default:
		if errors.Is(ctx.Err(), context.Canceled) {
			return nil
		}
		return ctx.Err()
	}
}

func (r *Runner) handleMessage(parent context.Context, msg InboundMessage) {
	log := r.logger.With(
		slog.String("transport", msg.Transport),
		slog.String("sender", msg.Sender),
		slog.String("thread", msg.ThreadID),
	)

	if len(r.allowedSenders) > 0 {
		if _, ok := r.allowedSenders[strings.ToLower(msg.Sender)]; !ok {
			log.Warn("sender not allowed")
			return
		}
	}

	reqCtx := parent
	if r.reqTimeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(parent, r.reqTimeout)
		defer cancel()
	}

	req := AgentRequest{
		Prompt:     msg.Text,
		History:    nil,
		Actions:    r.actionSpecs,
		SenderMeta: msg.Meta,
	}

	start := time.Now()
	resp, err := r.callAgentWithRetry(reqCtx, req, log)
	if err != nil {
		log.Error("agent error", slog.String("err", err.Error()))
		return
	}
	log.Info("agent reply", slog.Duration("ms", time.Since(start)))

	// Execute actions if any
	var actionResults []string
	for _, call := range resp.ActionCalls {
		if len(r.allowedActions) > 0 {
			if _, ok := r.allowedActions[call.Name]; !ok {
				log.Warn("action not allowed", slog.String("action", call.Name))
				r.logAudit(call.Name, msg.Sender, "denied", 0)
				continue
			}
		}
		act, ok := r.actions[call.Name]
		if !ok {
			log.Warn("unknown action", slog.String("action", call.Name))
			continue
		}
		aCtx := reqCtx
		if r.actionTimeout > 0 {
			var cancel context.CancelFunc
			aCtx, cancel = context.WithTimeout(reqCtx, r.actionTimeout)
			defer cancel()
		}
		aStart := time.Now()
		out, err := act.Invoke(aCtx, call.Args)
		if err != nil {
			log.Error("action error", slog.String("action", call.Name), slog.String("err", err.Error()))
			r.logAudit(call.Name, msg.Sender, "error", time.Since(aStart))
			continue
		}
		log.Info("action ok", slog.String("action", call.Name), slog.Duration("ms", time.Since(aStart)))
		r.logAudit(call.Name, msg.Sender, "ok", time.Since(aStart))
		if len(out) > 0 {
			actionResults = append(actionResults, fmt.Sprintf("[%s]\n%s", call.Name, string(out)))
		}
	}

	finalText := resp.Reply
	if len(actionResults) > 0 {
		finalText = finalText + "\n\n" + joinStrings(actionResults, "\n\n")
	}

	outMsg := OutboundMessage{
		Transport: msg.Transport,
		Recipient: msg.Sender,
		Text:      finalText,
		ThreadID:  msg.ThreadID,
	}

	tr, ok := r.transportMap[msg.Transport]
	if !ok {
		log.Error("no transport for outbound", slog.String("transport", msg.Transport))
		return
	}
	if err := r.sendWithRetry(reqCtx, tr, outMsg, log); err != nil {
		log.Error("send error", slog.String("err", err.Error()))
	}
}

func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	out := parts[0]
	for _, p := range parts[1:] {
		out += sep + p
	}
	return out
}

func (r *Runner) callAgentWithRetry(ctx context.Context, req AgentRequest, log *slog.Logger) (AgentResponse, error) {
	var resp AgentResponse
	var agentErr error
	err := retry(ctx, 3, func() error {
		var err error
		resp, err = r.agent.Generate(ctx, req)
		if err != nil {
			agentErr = err
			log.Warn("agent retry", slog.String("err", err.Error()))
		}
		return err
	})
	if err != nil {
		return resp, agentErr
	}
	return resp, nil
}

func (r *Runner) sendWithRetry(ctx context.Context, tr Transport, msg OutboundMessage, log *slog.Logger) error {
	var sendErr error
	err := retry(ctx, 3, func() error {
		err := tr.Send(ctx, msg)
		if err != nil {
			sendErr = err
			log.Warn("send retry", slog.String("err", err.Error()))
		}
		return err
	})
	if err != nil {
		return sendErr
	}
	return nil
}

func (r *Runner) logAudit(action, sender, outcome string, dur time.Duration) {
	if r.auditStore == nil {
		return
	}
	_ = r.auditStore.AppendAudit(action, sender, outcome, dur)
}
