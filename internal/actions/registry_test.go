package action

import (
	"context"
	"encoding/json"
	"testing"

	"nostr-codex-runner/internal/core"
)

type fakeAction struct{ name string }

func (f *fakeAction) Name() string           { return f.name }
func (f *fakeAction) Capabilities() []string { return nil }
func (f *fakeAction) Invoke(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	return json.RawMessage(`"ok"`), nil
}

func TestActionRegistry(t *testing.T) {
	t.Cleanup(func() {
		registryMu.Lock()
		registry = make(map[string]Constructor)
		registryMu.Unlock()
	})

	if err := Register("fake", func(cfg any) (core.Action, error) { return &fakeAction{name: "fake"}, nil }); err != nil {
		t.Fatalf("register err: %v", err)
	}
	act, err := Build("fake", nil)
	if err != nil {
		t.Fatalf("build err: %v", err)
	}
	if act.Name() != "fake" {
		t.Fatalf("unexpected name %s", act.Name())
	}
}

func TestActionRegistryDuplicate(t *testing.T) {
	t.Cleanup(func() {
		registryMu.Lock()
		registry = make(map[string]Constructor)
		registryMu.Unlock()
	})
	_ = Register("dup", func(cfg any) (core.Action, error) { return &fakeAction{name: "dup"}, nil })
	if err := Register("dup", nil); err == nil {
		t.Fatalf("expected duplicate error")
	}
}
