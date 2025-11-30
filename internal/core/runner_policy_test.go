package core

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

// additional tests for policy paths

type denyAction struct{ name string }

func (d *denyAction) Name() string           { return d.name }
func (d *denyAction) Capabilities() []string { return nil }
func (d *denyAction) Invoke(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	return json.RawMessage(`"ok"`), nil
}

type auditRecorder struct{ entries []string }

func (a *auditRecorder) AppendAudit(action, sender, outcome string, dur time.Duration) error {
	a.entries = append(a.entries, action+":"+outcome)
	return nil
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

	waitForChannel(t, &tr.inbound)
	tr.inbound <- InboundMessage{Transport: "mock", Sender: "alice", Text: "run", ThreadID: "t"}
	time.Sleep(20 * time.Millisecond)
	cancel()
	<-done

	if len(tr.sent) != 1 {
		t.Fatalf("expected 1 outbound")
	}
	if len(audit.entries) == 0 {
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

	waitForChannel(t, &tr.inbound)
	tr.inbound <- InboundMessage{Transport: "mock", Sender: "alice", Text: "run", ThreadID: "t"}
	time.Sleep(10 * time.Millisecond)
	cancel()
	<-done

	if len(tr.sent) != 0 {
		t.Fatalf("expected no outbound for disallowed sender")
	}
}
