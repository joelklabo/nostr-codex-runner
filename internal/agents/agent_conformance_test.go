package agent_test

import (
	"context"
	"testing"

	"nostr-codex-runner/internal/agents/echo"
	"nostr-codex-runner/internal/core"
)

func TestAgentConformanceSimple(t *testing.T) {
	ag := echo.New()
	resp, err := ag.Generate(context.Background(), core.AgentRequest{Prompt: "ping"})
	if err != nil {
		t.Fatalf("agent %T failed: %v", ag, err)
	}
	if resp.Reply == "" {
		t.Fatalf("agent %T returned empty reply", ag)
	}
}
