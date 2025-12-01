package codexcli

import (
	"context"
	"testing"

	"nostr-codex-runner/internal/core"
)

func TestGenerateEmptyPrompt(t *testing.T) {
	ag := New(Config{})
	if _, err := ag.Generate(context.Background(), core.AgentRequest{}); err == nil {
		t.Fatalf("expected error on empty prompt")
	}
}
