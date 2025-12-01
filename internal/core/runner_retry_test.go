package core

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"
)

type flakyAgent struct{ attempts int }

func (f *flakyAgent) Generate(ctx context.Context, req AgentRequest) (AgentResponse, error) {
	f.attempts++
	if f.attempts < 2 {
		return AgentResponse{}, errors.New("try again")
	}
	return AgentResponse{Reply: "ok"}, nil
}

type flakyTransport struct{ attempts int }

func (f *flakyTransport) ID() string { return "flaky" }
func (f *flakyTransport) Start(ctx context.Context, inbound chan<- InboundMessage) error {
	return nil
}
func (f *flakyTransport) Send(ctx context.Context, msg OutboundMessage) error {
	f.attempts++
	if f.attempts < 2 {
		return errors.New("fail send")
	}
	return nil
}

func TestCallAgentWithRetrySucceedsAfterRetry(t *testing.T) {
	r := &Runner{agent: &flakyAgent{}}
	log := slog.Default()
	resp, err := r.callAgentWithRetry(context.Background(), AgentRequest{Prompt: "hi"}, log)
	if err != nil {
		t.Fatalf("callAgentWithRetry err: %v", err)
	}
	if resp.Reply != "ok" {
		t.Fatalf("unexpected reply %s", resp.Reply)
	}
}

func TestSendWithRetryRetriesTransport(t *testing.T) {
	r := &Runner{}
	ft := &flakyTransport{}
	log := slog.Default()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := r.sendWithRetry(ctx, ft, OutboundMessage{Recipient: "a"}, log); err != nil {
		t.Fatalf("sendWithRetry err: %v", err)
	}
	if ft.attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", ft.attempts)
	}
}
