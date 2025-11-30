package shell

import (
	"context"
	"encoding/json"
	"testing"
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
