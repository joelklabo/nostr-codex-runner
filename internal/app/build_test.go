package app

import (
	"context"
	"testing"
	"time"

	"nostr-codex-runner/internal/config"
	"nostr-codex-runner/internal/store"
)

func TestBuildWithMockEcho(t *testing.T) {
	cfg := &config.Config{
		Transports: []config.TransportConfig{{Type: "mock", ID: "mock"}},
		Agent:      config.AgentConfig{Type: "echo"},
		Actions:    []config.ActionConfig{},
		Runner:     config.RunnerConfig{AllowedPubkeys: []string{"mock"}},
		Storage:    config.StorageConfig{Path: t.TempDir() + "/state.db"},
		Logging:    config.LoggingConfig{Level: "error"},
	}
	st, err := store.New(cfg.Storage.Path)
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Errorf("close store: %v", err)
		}
	})

	r, err := Build(cfg, st, nil)
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() { errCh <- r.Start(ctx) }()

	// let it start then cancel
	time.Sleep(20 * time.Millisecond)
	cancel()

	if err := <-errCh; err != nil && err != context.Canceled {
		t.Fatalf("runner err: %v", err)
	}
}
