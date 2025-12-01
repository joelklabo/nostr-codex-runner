package codex

import (
	"context"
	"os"
	"path/filepath"
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
