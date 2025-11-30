package agent

import (
	"context"
	"testing"

	"nostr-codex-runner/internal/core"
)

type fakeAgent struct{}

func (f *fakeAgent) Generate(ctx context.Context, req core.AgentRequest) (core.AgentResponse, error) {
	return core.AgentResponse{Reply: "ok"}, nil
}

func TestAgentRegistry(t *testing.T) {
	t.Cleanup(func() {
		registryMu.Lock()
		registry = make(map[string]Constructor)
		registryMu.Unlock()
	})

	if err := Register("fake", func(cfg any) (core.Agent, error) { return &fakeAgent{}, nil }); err != nil {
		t.Fatalf("register err: %v", err)
	}
	ag, err := Build("fake", nil)
	if err != nil {
		t.Fatalf("build err: %v", err)
	}
	if _, err := ag.Generate(context.Background(), core.AgentRequest{Prompt: "hi"}); err != nil {
		t.Fatalf("generate err: %v", err)
	}
}

func TestAgentRegistryDuplicate(t *testing.T) {
	t.Cleanup(func() {
		registryMu.Lock()
		registry = make(map[string]Constructor)
		registryMu.Unlock()
	})
	_ = Register("dup", func(cfg any) (core.Agent, error) { return &fakeAgent{}, nil })
	if err := Register("dup", nil); err == nil {
		t.Fatalf("expected duplicate error")
	}
}
