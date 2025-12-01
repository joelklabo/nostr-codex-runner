package app

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"nostr-codex-runner/internal/config"
	"nostr-codex-runner/internal/core"
	"nostr-codex-runner/internal/store"
	tmock "nostr-codex-runner/internal/transports/mock"
)

// End-to-end smoke using mock transport + echo agent to ensure wiring works.
func TestE2EFlowWithMockTransport(t *testing.T) {
	cfg := &config.Config{
		Runner:     config.RunnerConfig{PrivateKey: "k", AllowedPubkeys: []string{"alice"}},
		Storage:    config.StorageConfig{Path: t.TempDir() + "/state.db"},
		Transports: []config.TransportConfig{{Type: "mock", ID: "mock"}},
		Agent:      config.AgentConfig{Type: "echo"},
		Actions:    []config.ActionConfig{{Type: "shell", Name: "shell", Workdir: "."}},
		Projects:   []config.Project{{ID: "p", Name: "p", Path: "."}},
	}
	st, err := store.New(cfg.Storage.Path)
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	defer func() { _ = st.Close() }()

	r, err := Build(cfg, st, slog.Default())
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	// Grab the mock transport to inject a message.
	var mt *tmock.Transport
	for _, tr := range r.Transports() {
		if candidate, ok := tr.(*tmock.Transport); ok {
			mt = candidate
			break
		}
	}
	if mt == nil {
		t.Fatalf("mock transport not found")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = r.Start(ctx) }()

	inbound := mt.Inbound
	outbound := mt.Outbound

	inbound <- core.InboundMessage{Transport: mt.ID(), Sender: "alice", Text: "hello", ThreadID: "t"}

	select {
	case msg := <-outbound:
		if msg.Text == "" {
			t.Fatalf("empty outbound")
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for outbound")
	}
}
