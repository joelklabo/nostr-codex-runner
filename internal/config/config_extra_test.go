package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

func TestNormalizePubkeyHandlesNpub(t *testing.T) {
	priv := nostr.GeneratePrivateKey()
	pub, _ := nostr.GetPublicKey(priv)
	npub, _ := nip19.EncodePublicKey(pub)
	if got := normalizePubkey(npub); got != pub {
		t.Fatalf("normalizePubkey failed: %s", got)
	}
}

func TestApplyDefaultsFillsTransportAgentAndProjects(t *testing.T) {
	priv := nostr.GeneratePrivateKey()
	pub, _ := nostr.GetPublicKey(priv)
	cfg := Config{
		Runner:  RunnerConfig{PrivateKey: priv, AllowedPubkeys: []string{pub}},
		Actions: []ActionConfig{{Type: "shell"}},
	}
	cfg.applyDefaults(".")
	if len(cfg.Transports) == 0 || cfg.Transports[0].Type != "nostr" {
		t.Fatalf("expected nostr default transport")
	}
	if cfg.Agent.Type != "codexcli" {
		t.Fatalf("expected default agent type, got %s", cfg.Agent.Type)
	}
	if len(cfg.Projects) == 0 || cfg.Projects[0].Path == "" {
		t.Fatalf("projects default missing")
	}
	if cfg.Runner.AllowedPubkeys[0] != pub {
		t.Fatalf("allowed pubkey not normalized: %v", cfg.Runner.AllowedPubkeys)
	}
}

func TestExpandPathEnvAndHome(t *testing.T) {
	os.Setenv("NCR_TMP", "/tmp/ncrpath")
	defer os.Unsetenv("NCR_TMP")
	got := expandPath("$NCR_TMP/sub")
	if got != filepath.Clean("/tmp/ncrpath/sub") {
		t.Fatalf("expand env failed: %s", got)
	}
	home, _ := os.UserHomeDir()
	if expandPath("~/x") != filepath.Join(home, "x") {
		t.Fatalf("expand home failed")
	}
}

func TestValidateProjectPath(t *testing.T) {
	cfg := Config{
		Runner:  RunnerConfig{PrivateKey: "k", AllowedPubkeys: []string{"a"}},
		Storage: StorageConfig{Path: "/tmp/state.db"},
		Transports: []TransportConfig{{
			Type:           "mock",
			PrivateKey:     "k",
			AllowedPubkeys: []string{"a"},
		}},
		Agent:    AgentConfig{Type: "echo"},
		Actions:  []ActionConfig{{Type: "shell"}},
		Projects: []Project{{ID: "p1", Name: "p1", Path: ""}},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for empty project path")
	}
}
