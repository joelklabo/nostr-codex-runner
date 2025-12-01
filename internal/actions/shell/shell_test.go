package shell

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestAllowedCommand(t *testing.T) {
	act := New(Config{Allowed: []string{"echo"}})
	resp, err := act.Invoke(context.Background(), json.RawMessage(`{"command":"echo hi"}`))
	if err != nil {
		t.Fatalf("invoke err: %v", err)
	}
	if len(resp) == 0 {
		t.Fatalf("empty response")
	}
}

func TestDeniedCommand(t *testing.T) {
	act := New(Config{Allowed: []string{"echo"}})
	_, err := act.Invoke(context.Background(), json.RawMessage(`"rm -rf /"`))
	if err == nil {
		t.Fatalf("expected deny error")
	}
}

func TestMissingCommand(t *testing.T) {
	act := New(Config{})
	if _, err := act.Invoke(context.Background(), nil); err == nil {
		t.Fatalf("expected missing command error")
	}
	if _, err := act.Invoke(context.Background(), json.RawMessage(`""`)); err == nil {
		t.Fatalf("expected empty command error")
	}
}

func TestTimeoutAndExitCode(t *testing.T) {
	act := New(Config{TimeoutSeconds: 1})
	start := time.Now()
	if _, err := act.Invoke(context.Background(), json.RawMessage(`"sleep 2"`)); err == nil {
		t.Fatalf("expected timeout")
	}
	if time.Since(start) > 3*time.Second {
		t.Fatalf("timeout not enforced")
	}

	resp, err := act.Invoke(context.Background(), json.RawMessage(`"bash -lc 'exit 3'"`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(resp), "exit=3") {
		t.Fatalf("expected exit code in response: %s", string(resp))
	}
}

func TestTruncateAndHelp(t *testing.T) {
	if help := New(Config{}).Help(); help == "" {
		t.Fatalf("help should not be empty")
	}
	if got := truncate("abcdef", 3); got != "abc" {
		t.Fatalf("truncate mismatch %s", got)
	}
	if got := truncate("abcdef", 5); got != "ab..." {
		t.Fatalf("truncate ellipsis mismatch %s", got)
	}
}
