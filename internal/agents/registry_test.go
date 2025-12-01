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
	if got := RegisteredTypes(); len(got) != 1 || got[0] != "fake" {
		t.Fatalf("registered types mismatch: %v", got)
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

func TestAgentMustRegisterPanics(t *testing.T) {
	t.Cleanup(func() {
		registryMu.Lock()
		registry = make(map[string]Constructor)
		registryMu.Unlock()
	})
	MustRegister("x", func(cfg any) (core.Agent, error) { return &fakeAgent{}, nil })
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()
	MustRegister("x", nil)
}
