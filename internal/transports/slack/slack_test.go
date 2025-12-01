package slack

import (
	"context"
	"testing"

	"nostr-codex-runner/internal/core"
)

func TestSlackIDAndSend(t *testing.T) {
	tr := New(Config{})
	if tr.ID() != "slack" {
		t.Fatalf("id mismatch %s", tr.ID())
	}
	if err := tr.Send(context.Background(), core.OutboundMessage{}); err == nil {
		t.Fatalf("expected not implemented error")
	}
}

func TestSlackStartReturnsContextError(t *testing.T) {
	tr := New(Config{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := tr.Start(ctx, make(chan core.InboundMessage)); err == nil {
		t.Fatalf("expected context error")
	}
}
