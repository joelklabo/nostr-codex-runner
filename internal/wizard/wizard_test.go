package wizard

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesConfig(t *testing.T) {
	td := t.TempDir()
	path := filepath.Join(td, "config.yaml")
	p := &StubPrompter{
		Selects:   []string{"mock-echo"},
		Inputs:    []string{}, // no nostr prompts for mock preset
		Passwords: []string{"abcd1234"},
		Confirms:  []bool{true, false}, // overwrite? dry-run?
	}
	got, err := Run(context.Background(), path, p)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if got != path {
		t.Fatalf("expected path %s, got %s", path, got)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "type: mock") {
		t.Fatalf("config missing expected fields:\n%s", content)
	}
}

func TestRunRequiresAllowedPubkey(t *testing.T) {
	td := t.TempDir()
	path := filepath.Join(td, "config.yaml")
	p := &StubPrompter{
		Selects:   []string{"claude-dm"},
		Inputs:    []string{"", ""}, // relays blank, allowed blank
		Passwords: []string{"abcd1234"},
		Confirms:  []bool{false}, // overwrite?
	}
	_, err := Run(context.Background(), path, p)
	if err == nil {
		t.Fatalf("expected error for missing allowed pubkeys")
	}
}

func containsAll(s string, needles []string) bool {
	for _, n := range needles {
		if !strings.Contains(s, n) {
			return false
		}
	}
	return true
}
