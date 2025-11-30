package transport

import (
	"context"
	"testing"

	"nostr-codex-runner/internal/core"
)

type fakeTransport struct{ id string }

func (f *fakeTransport) ID() string { return f.id }
func (f *fakeTransport) Start(ctx context.Context, inbound chan<- core.InboundMessage) error {
	return nil
}
func (f *fakeTransport) Send(ctx context.Context, msg core.OutboundMessage) error { return nil }

func TestRegistryRegistersAndBuilds(t *testing.T) {
	t.Cleanup(func() {
		// reset registry for isolation
		registryMu.Lock()
		registry = make(map[string]Constructor)
		registryMu.Unlock()
	})

	err := Register("fake", func(cfg any) (core.Transport, error) { return &fakeTransport{id: "x"}, nil })
	if err != nil {
		t.Fatalf("register err: %v", err)
	}
	tr, err := Build("fake", nil)
	if err != nil {
		t.Fatalf("build err: %v", err)
	}
	if tr.ID() != "x" {
		t.Fatalf("unexpected id %s", tr.ID())
	}
}

func TestRegistryDuplicate(t *testing.T) {
	t.Cleanup(func() {
		registryMu.Lock()
		registry = make(map[string]Constructor)
		registryMu.Unlock()
	})

	_ = Register("dup", func(cfg any) (core.Transport, error) { return &fakeTransport{id: "a"}, nil })
	if err := Register("dup", nil); err == nil {
		t.Fatalf("expected duplicate error")
	}
}
