package core

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"testing"
	"time"
)

type mockTransport struct {
	id      string
	mu      sync.Mutex
	inbound chan<- InboundMessage
	sent    []OutboundMessage
}

func (m *mockTransport) ID() string { return m.id }

func (m *mockTransport) Start(ctx context.Context, in chan<- InboundMessage) error {
	m.setInbound(in)
	<-ctx.Done()
	return ctx.Err()
}

func (m *mockTransport) Send(_ context.Context, msg OutboundMessage) error {
	m.appendSent(msg)
	return nil
}

func (m *mockTransport) setInbound(ch chan<- InboundMessage) {
	m.mu.Lock()
	m.inbound = ch
	m.mu.Unlock()
}

func (m *mockTransport) inboundChan() chan<- InboundMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.inbound
}

func (m *mockTransport) appendSent(msg OutboundMessage) {
	m.mu.Lock()
	m.sent = append(m.sent, msg)
	m.mu.Unlock()
}

func (m *mockTransport) sentMessages() []OutboundMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]OutboundMessage, len(m.sent))
	copy(out, m.sent)
	return out
}

type mockAgent struct {
	reply       string
	calls       []AgentRequest
	actionCalls []ActionCall
}

func (m *mockAgent) Generate(ctx context.Context, req AgentRequest) (AgentResponse, error) {
	m.calls = append(m.calls, req)
	return AgentResponse{Reply: m.reply, ActionCalls: m.actionCalls}, nil
}

type mockAction struct {
	name   string
	result string
}

func (m *mockAction) Name() string           { return m.name }
func (m *mockAction) Capabilities() []string { return nil }
func (m *mockAction) Help() string           { return "" }
func (m *mockAction) Invoke(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	return json.RawMessage(m.result), nil
}

func TestRunnerEndToEnd(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tr := &mockTransport{id: "mock"}
	ag := &mockAgent{reply: "hi"}
	sh := &mockAction{name: "echo", result: "pong"}

	r := NewRunner([]Transport{tr}, ag, []Action{sh}, slog.Default(), WithRequestTimeout(2*time.Second), WithActionTimeout(2*time.Second))

	done := make(chan struct{})
	go func() {
		if err := r.Start(ctx); err != nil && err != context.Canceled {
			t.Errorf("runner err: %v", err)
		}
		close(done)
	}()

	inCh := waitForChannel(t, tr.inboundChan)
	inbound := InboundMessage{Transport: "mock", Sender: "alice", Text: "hello", ThreadID: "t1"}
	inCh <- inbound

	// allow processing
	time.Sleep(50 * time.Millisecond)
	cancel()
	<-done

	sent := tr.sentMessages()
	if len(sent) != 1 {
		t.Fatalf("expected 1 outbound, got %d", len(sent))
	}
	if sent[0].Recipient != "alice" {
		t.Fatalf("recipient mismatch: %s", sent[0].Recipient)
	}
	if sent[0].Text == "" {
		t.Fatalf("empty outbound text")
	}
}

func TestRunnerExecutesActionResults(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tr := &mockTransport{id: "mock"}
	ag := &mockAgent{reply: "base", actionCalls: []ActionCall{{Name: "echo"}}}
	sh := &mockAction{name: "echo", result: "ok"}

	r := NewRunner([]Transport{tr}, ag, []Action{sh}, slog.Default())

	done := make(chan struct{})
	go func() {
		_ = r.Start(ctx)
		close(done)
	}()

	inCh := waitForChannel(t, tr.inboundChan)
	inCh <- InboundMessage{Transport: "mock", Sender: "bob", Text: "run", ThreadID: "th"}
	time.Sleep(50 * time.Millisecond)
	cancel()
	<-done

	sent := tr.sentMessages()
	if len(sent) != 1 {
		t.Fatalf("expected 1 outbound, got %d", len(sent))
	}
	if got := sent[0].Text; got == "base" {
		t.Fatalf("expected action result appended, got %q", got)
	}
}

func waitForChannel(t *testing.T, chFn func() chan<- InboundMessage) chan<- InboundMessage {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if ch := chFn(); ch != nil {
			return ch
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("transport inbound channel not set")
	return nil
}
