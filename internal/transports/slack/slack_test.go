package slack

import (
	"context"
	"testing"

	"nostr-codex-runner/internal/core"
)

func TestSlackIDAndSend(t *testing.T) {
	tr := New(Config{})
	if tr.ID() != "slack" {
		t.Fatalf("id mismatch")
	}
	if err := tr.Send(context.Background(), core.OutboundMessage{}); err == nil {
		t.Fatalf("expected not implemented error")
	}
}
