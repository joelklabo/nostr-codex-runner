package copilotcli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nostr-codex-runner/internal/core"
)

// create a fake copilot binary that returns static text.
func writeFakeCopilot(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "copilot")
	script := "#!/usr/bin/env bash\necho \"hi from copilot\"\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake gh: %v", err)
	}
	return path
}

func TestCopilotAgentGenerates(t *testing.T) {
	td := t.TempDir()
	bin := writeFakeCopilot(t, td)

	ag := New(Config{Binary: bin})
	resp, err := ag.Generate(context.Background(), core.AgentRequest{Prompt: "hello"})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if resp.Reply == "" {
		t.Fatalf("empty reply")
	}
}

func TestCopilotAgentTimeout(t *testing.T) {
	td := t.TempDir()
	bin := filepath.Join(td, "slowcopilot")
	script := "#!/usr/bin/env bash\nsleep 2\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	ag := New(Config{Binary: bin, TimeoutSeconds: 1})
	_, err := ag.Generate(context.Background(), core.AgentRequest{Prompt: "hi"})
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("expected timeout error, got %v", err)
	}
}

func TestCopilotAgentAllowAllTools(t *testing.T) {
	td := t.TempDir()
	bin := filepath.Join(td, "args")
	script := "#!/usr/bin/env bash\necho \"$@\"\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	ag := New(Config{
		Binary:         bin,
		AllowAllTools:  true,
		ExtraArgs:      []string{"--foo", "bar"},
		TimeoutSeconds: 2,
	})
	resp, err := ag.Generate(context.Background(), core.AgentRequest{Prompt: "prompt"})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if !strings.Contains(resp.Reply, "--allow-all-tools") || !strings.Contains(resp.Reply, "--foo") {
		t.Fatalf("expected flags in reply: %s", resp.Reply)
	}
	if !strings.Contains(resp.Reply, "prompt") {
		t.Fatalf("prompt missing from args: %s", resp.Reply)
	}
	time.Sleep(10 * time.Millisecond) // ensure script finished
}
