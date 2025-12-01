package app

import (
	"log/slog"
	"path/filepath"
	"testing"

	"nostr-codex-runner/internal/config"
	"nostr-codex-runner/internal/store"
)

func TestBuildFailsOnUnknownAgent(t *testing.T) {
	cfg := &config.Config{
		Runner: config.RunnerConfig{PrivateKey: "k", AllowedPubkeys: []string{"a"}},
		Storage: config.StorageConfig{
			Path: filepath.Join(t.TempDir(), "state.db"),
		},
		Transports: []config.TransportConfig{{Type: "mock", ID: "mock"}},
		Agent:      config.AgentConfig{Type: "nope"},
		Actions:    []config.ActionConfig{{Type: "shell"}},
		Projects:   []config.Project{{ID: "p", Name: "p", Path: "."}},
	}
	st, _ := store.New(cfg.Storage.Path)
	defer st.Close()
	if _, err := Build(cfg, st, slog.Default()); err == nil {
		t.Fatalf("expected error for unknown agent")
	}
}

func TestBuildFailsOnUnknownAction(t *testing.T) {
	cfg := &config.Config{
		Runner: config.RunnerConfig{PrivateKey: "k", AllowedPubkeys: []string{"a"}},
		Storage: config.StorageConfig{
			Path: filepath.Join(t.TempDir(), "state.db"),
		},
		Transports: []config.TransportConfig{{Type: "mock", ID: "mock"}},
		Agent:      config.AgentConfig{Type: "echo"},
		Actions:    []config.ActionConfig{{Type: "nope"}},
		Projects:   []config.Project{{ID: "p", Name: "p", Path: "."}},
	}
	st, _ := store.New(cfg.Storage.Path)
	defer st.Close()
	if _, err := Build(cfg, st, slog.Default()); err == nil {
		t.Fatalf("expected error for unknown action")
	}
}
