package codex

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nostr-codex-runner/internal/config"
)

func TestParseCodexJSONLSuccess(t *testing.T) {
	lines := `{"type":"message","thread_id":"t1"}
{"type":"message","item":{"type":"agent_message","text":"hello"}}
`
	res, err := parseCodexJSONL([]byte(lines))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if res.SessionID != "t1" {
		t.Fatalf("session id %s", res.SessionID)
	}
	if res.Reply != "hello" {
		t.Fatalf("reply %s", res.Reply)
	}
}

func TestParseCodexJSONLError(t *testing.T) {
	lines := `{"type":"message","thread_id":"t1","error":"boom"}`
	if _, err := parseCodexJSONL([]byte(lines)); err == nil {
		t.Fatalf("expected error")
	}
}

func TestContextWithTimeout(t *testing.T) {
	r := New(config.CodexConfig{TimeoutSeconds: 1})
	ctx, cancel := r.ContextWithTimeout(context.Background())
	defer cancel()
	select {
	case <-time.After(1500 * time.Millisecond):
		// context should expire within timeout window
	case <-ctx.Done():
		// ok
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()
	got := expandPath("~/x")
	want := filepath.Join(home, "x")
	if got != want {
		t.Fatalf("expandPath got %s want %s", got, want)
	}
}

func TestRunValidatesPrompt(t *testing.T) {
	r := New(config.CodexConfig{Binary: "echo"})
	if _, err := r.Run(context.Background(), "", "   "); err == nil {
		t.Fatalf("expected prompt validation error")
	}
}

func TestRunExecutesBinary(t *testing.T) {
	td := t.TempDir()
	bin := filepath.Join(td, "codexfake")
	script := `#!/usr/bin/env bash
echo '{"thread_id":"s1","type":"message"}'
echo '{"item":{"type":"agent_message","text":"hi"}}'
`
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	r := New(config.CodexConfig{Binary: bin, TimeoutSeconds: 1})
	res, err := r.Run(context.Background(), "", "hello")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if res.SessionID != "s1" || res.Reply != "hi" {
		t.Fatalf("unexpected result %+v", res)
	}
}

func TestRunCapturesExecError(t *testing.T) {
	td := t.TempDir()
	bin := filepath.Join(td, "codexfail")
	script := "#!/usr/bin/env bash\necho oops 1>&2\nexit 1\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	r := New(config.CodexConfig{Binary: bin})
	_, err := r.Run(context.Background(), "", "hello")
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitErr, ok := err.(*exec.Error); ok && exitErr.Err == exec.ErrNotFound {
		t.Fatalf("script not executable: %v", err)
	}
	if err != nil && !strings.Contains(err.Error(), "codex exec failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseCodexJSONLMissingSession(t *testing.T) {
	lines := `{"type":"message","item":{"type":"agent_message","text":"hi"}}`
	if _, err := parseCodexJSONL([]byte(lines)); err == nil {
		t.Fatalf("expected error when session id missing")
	}
}

func TestRunResumeAddsSessionID(t *testing.T) {
	td := t.TempDir()
	argsFile := filepath.Join(td, "args.txt")
	bin := filepath.Join(td, "codexresume")
	script := "#!/usr/bin/env bash\necho \"$@\" > " + argsFile + "\necho '{\"thread_id\":\"s1\"}'\necho '{\"item\":{\"type\":\"agent_message\",\"text\":\"ok\"}}'\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	r := New(config.CodexConfig{Binary: bin})
	if _, err := r.Run(context.Background(), "session-123", "hello"); err != nil {
		t.Fatalf("run: %v", err)
	}
	data, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("read args: %v", err)
	}
	argStr := string(data)
	if !strings.Contains(argStr, "resume") || !strings.Contains(argStr, "session-123") {
		t.Fatalf("expected resume args, got %s", argStr)
	}
}
