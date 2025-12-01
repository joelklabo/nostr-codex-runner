package app

import (
	"log/slog"
	"path/filepath"
	"testing"

	"nostr-codex-runner/internal/config"
	"nostr-codex-runner/internal/store"
)

func TestBuildWithWhatsAppAndCopilot(t *testing.T) {
	td := t.TempDir()
	cfg := &config.Config{
		Runner: config.RunnerConfig{
			PrivateKey:     "abcd",
			AllowedPubkeys: []string{"1234"},
			MaxReplyChars:  4000,
		},
		Storage: config.StorageConfig{Path: filepath.Join(td, "state.db")},
		Transports: []config.TransportConfig{
			{
				Type: "whatsapp",
				ID:   "wa",
				Config: map[string]any{
					"account_sid": "AC123",
					"auth_token":  "token",
					"from_number": "whatsapp:+15550001234",
					"listen":      "127.0.0.1:0",
					"path":        "/hook",
				},
			},
		},
		Agent: config.AgentConfig{
			Type:   "copilotcli",
			Config: config.CodexConfig{Binary: "echo", TimeoutSeconds: 1},
		},
		Actions: []config.ActionConfig{
			{Type: "readfile", Name: "read", Roots: []string{td}, MaxBytes: 2048},
			{Type: "writefile", Name: "write", Roots: []string{td}, AllowWrite: true, MaxBytes: 2048},
		},
		Projects: []config.Project{{ID: "default", Name: "default", Path: "."}},
	}
	st, err := store.New(cfg.Storage.Path)
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	defer st.Close()

	r, err := Build(cfg, st, slog.Default())
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if r == nil {
		t.Fatalf("runner is nil")
	}
}

func TestBuildWhatsAppConfigError(t *testing.T) {
	td := t.TempDir()
	cfg := &config.Config{
		Runner:  config.RunnerConfig{PrivateKey: "abcd", AllowedPubkeys: []string{"1234"}},
		Storage: config.StorageConfig{Path: filepath.Join(td, "state.db")},
		Transports: []config.TransportConfig{
			{
				Type: "whatsapp",
				ID:   "wa",
				Config: map[string]any{
					"account_sid": make(chan int),
					"auth_token":  "token",
					"from_number": "whatsapp:+1",
				},
			},
		},
		Agent:    config.AgentConfig{Type: "echo"},
		Actions:  []config.ActionConfig{{Type: "shell", Name: "shell"}},
		Projects: []config.Project{{ID: "p", Name: "p", Path: "."}},
	}
	st, _ := store.New(cfg.Storage.Path)
	defer st.Close()

	if _, err := Build(cfg, st, slog.Default()); err == nil {
		t.Fatalf("expected decode error")
	}
}
