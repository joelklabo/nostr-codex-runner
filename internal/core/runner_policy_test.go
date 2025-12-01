package core

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"
)

// additional tests for policy paths

type denyAction struct{ name string }

func (d *denyAction) Name() string           { return d.name }
func (d *denyAction) Capabilities() []string { return nil }
func (d *denyAction) Help() string           { return "" }
func (d *denyAction) Invoke(_ context.Context, args json.RawMessage) (json.RawMessage, error) {
	return json.RawMessage(`"ok"`), nil
}

type auditRecorder struct {
	mu      sync.Mutex
	entries []string
}

func (a *auditRecorder) AppendAudit(action, sender, outcome string, dur time.Duration) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.entries = append(a.entries, action+":"+outcome)
	return nil
}

func (a *auditRecorder) snapshot() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]string, len(a.entries))
	copy(out, a.entries)
	return out
}

func TestRunnerRespectsAllowedActions(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tr := &mockTransport{id: "mock"}
	ag := &mockAgent{reply: "hi", actionCalls: []ActionCall{{Name: "deny"}}}
	act := &denyAction{name: "deny"}
	audit := &auditRecorder{}

	r := NewRunner([]Transport{tr}, ag, []Action{act}, nil, WithAllowedActions([]string{"other"}), WithAuditLogger(audit))

	done := make(chan struct{})
	go func() {
		_ = r.Start(ctx)
		close(done)
	}()

	inCh := waitForChannel(t, tr.inboundChan)
	inCh <- InboundMessage{Transport: "mock", Sender: "alice", Text: "run", ThreadID: "t"}
	time.Sleep(20 * time.Millisecond)
	cancel()
	<-done

	sent := tr.sentMessages()
	if len(sent) != 1 {
		t.Fatalf("expected 1 outbound")
	}
	if len(audit.snapshot()) == 0 {
		t.Fatalf("expected audit entry")
	}
}

func TestRunnerSenderAllowlist(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tr := &mockTransport{id: "mock"}
	ag := &mockAgent{reply: "hi"}

	r := NewRunner([]Transport{tr}, ag, nil, nil, WithAllowedSenders([]string{"bob"}))

	done := make(chan struct{})
	go func() {
		_ = r.Start(ctx)
		close(done)
	}()

	inCh := waitForChannel(t, tr.inboundChan)
	inCh <- InboundMessage{Transport: "mock", Sender: "alice", Text: "run", ThreadID: "t"}
	time.Sleep(10 * time.Millisecond)
	cancel()
	<-done

	if len(tr.sentMessages()) != 0 {
		t.Fatalf("expected no outbound for disallowed sender")
	}
}
