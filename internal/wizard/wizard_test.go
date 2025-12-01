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
		Confirms:  []bool{true, false, true}, // overwrite? dry-run? continue missing deps?
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

func TestRunFailsOnDepsMissingWhenUserSaysNo(t *testing.T) {
	td := t.TempDir()
	path := filepath.Join(td, "config.yaml")
	p := &StubPrompter{
		Selects:   []string{"copilot-shell"},
		Inputs:    []string{"wss://relay.example", "npub1"}, // relays, allowed
		Passwords: []string{"abcd1234"},                     // nostr priv
		Confirms:  []bool{true, false, false},               // overwrite? dry-run? continue deps?
	}
	_, err := Run(context.Background(), path, p)
	if err == nil {
		t.Fatalf("expected failure when declining missing deps")
	}
}

func TestSetRegistryOverridesOptions(t *testing.T) {
	orig := GetRegistry()
	defer SetRegistry(orig)

	custom := Registry{
		Presets: []PresetOption{{Name: "mock-echo", Description: "Only mock"}},
		Transports: []TransportOption{
			{Name: "mock", Description: "Offline mock transport"},
		},
		Agents: []AgentOption{
			{Name: "echo", Description: "Echo agent"},
		},
		Actions: []ActionOption{
			{Name: "readfile", Description: "Read files", DefaultEnable: true},
		},
	}
	SetRegistry(custom)

	td := t.TempDir()
	path := filepath.Join(td, "config.yaml")
	p := &StubPrompter{
		Selects:   []string{"mock-echo"},
		Inputs:    []string{}, // no nostr prompts
		Passwords: []string{"abcd1234"},
		Confirms:  []bool{true, false, true}, // overwrite? dry-run? continue deps?
	}
	if _, err := Run(context.Background(), path, p); err != nil {
		t.Fatalf("run with custom registry: %v", err)
	}
}

func TestRunDryRunDoesNotWriteFile(t *testing.T) {
	td := t.TempDir()
	path := filepath.Join(td, "config.yaml")
	p := &StubPrompter{
		Selects:   []string{"mock-echo"},
		Inputs:    []string{},
		Passwords: []string{"abcd1234"},
		Confirms:  []bool{true, true}, // overwrite? dry-run?
	}
	got, err := Run(context.Background(), path, p)
	if err != nil {
		t.Fatalf("run dry-run: %v", err)
	}
	if got != path {
		t.Fatalf("expected path %s, got %s", path, got)
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("config should not be written on dry-run")
	}
}
