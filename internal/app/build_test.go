package app

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"nostr-codex-runner/internal/config"
	"nostr-codex-runner/internal/store"
)

func TestBuildWithMockTransport(t *testing.T) {
	td := t.TempDir()
	cfg := &config.Config{
		Runner: config.RunnerConfig{
			PrivateKey:     "abcd",
			AllowedPubkeys: []string{"1234"},
			MaxReplyChars:  4000,
		},
		Storage: config.StorageConfig{Path: filepath.Join(td, "state.db")},
		Transports: []config.TransportConfig{
			{Type: "mock", ID: "mock1"},
		},
		Agent: config.AgentConfig{
			Type: "echo",
		},
		Actions: []config.ActionConfig{
			{Type: "shell", Name: "shell", Workdir: ".", TimeoutSecs: 5, MaxOutput: 1000},
		},
		Projects: []config.Project{{ID: "default", Name: "default", Path: "."}},
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
	if r == nil {
		t.Fatalf("runner is nil")
	}
}

func TestBuildUnknownTransport(t *testing.T) {
	cfg := &config.Config{
		Runner: config.RunnerConfig{PrivateKey: "abcd", AllowedPubkeys: []string{"1234"}, MaxReplyChars: 1000},
		Storage: config.StorageConfig{
			Path: filepath.Join(os.TempDir(), "state.db"),
		},
		Transports: []config.TransportConfig{{Type: "nope", ID: "x"}},
		Agent:      config.AgentConfig{Type: "echo"},
		Actions:    []config.ActionConfig{{Type: "shell", Name: "shell"}},
		Projects:   []config.Project{{ID: "default", Name: "default", Path: "."}},
	}
	st, _ := store.New(cfg.Storage.Path)
	defer func() { _ = st.Close() }()
	if _, err := Build(cfg, st, slog.Default()); err == nil {
		t.Fatalf("expected error for unknown transport")
	}
}
