package presets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetUsesHomeOverride(t *testing.T) {
	td := t.TempDir()
	t.Setenv("HOME", td)
	overrideDir := filepath.Join(td, ".config", "buddy", "presets")
	if err := os.MkdirAll(overrideDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	overridePath := filepath.Join(overrideDir, "mock-echo.yaml")
	custom := []byte("meta:\n  description: override\ntransports: []\nagent:\n  type: echo\nactions: []\nrunner:\n  allowed_pubkeys: []\n")
	if err := os.WriteFile(overridePath, custom, 0o644); err != nil {
		t.Fatalf("write override: %v", err)
	}

	got, err := Get("mock-echo")
	if err != nil {
		t.Fatalf("get override: %v", err)
	}
	if string(got) != string(custom) {
		t.Fatalf("expected override content, got: %s", string(got))
	}
}

func TestGetFallsBackToEmbedded(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	got, err := Get("claude-dm")
	if err != nil {
		t.Fatalf("get embedded: %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("expected embedded content")
	}
}

func TestGetUnknown(t *testing.T) {
	if _, err := Get("nope"); err == nil {
		t.Fatalf("expected error for unknown preset")
	}
}
