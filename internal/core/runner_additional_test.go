package core

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"nostr-codex-runner/internal/commands"
	"nostr-codex-runner/internal/store"
)

// memoryStoreWithTime embeds memoryStore to tweak timestamps for timeout checks.
type memoryStoreWithTime struct{ memoryStore }

func TestPreparePromptClearsExpiredSession(t *testing.T) {
	st := &memoryStoreWithTime{}
	old := time.Now().Add(-2 * time.Minute)
	st.active = map[string]store.SessionState{"alice": {SessionID: "old", UpdatedAt: old}}
	r := NewRunner(nil, &mockAgent{reply: "hi"}, nil, slog.Default(), WithStore(st), WithSessionTimeout(time.Minute))
	_, sess := r.preparePrompt(commands.Command{Name: "run", Args: "prompt", Raw: "prompt"}, "alice")
	if sess != "" {
		t.Fatalf("expected session to be cleared, got %s", sess)
	}
	if _, ok := st.active["alice"]; ok {
		t.Fatalf("session not cleared from store")
	}
}

func TestInitialPromptPrepended(t *testing.T) {
	ag := &mockAgent{reply: "hi"}
	r := NewRunner(nil, ag, nil, slog.Default(), WithInitialPrompt("init"))
	outCh := make(chan OutboundMessage, 1)
	r.transportMap = map[string]Transport{"mock": &transportSpy{out: outCh}}

	msg := InboundMessage{Transport: "mock", Sender: "alice", Text: "hello", ThreadID: "t"}
	r.handleMessage(context.Background(), msg)

	if len(ag.calls) == 0 || ag.calls[0].Prompt != "init\n\nhello" {
		t.Fatalf("initial prompt not prepended, got %+v", ag.calls)
	}
	select {
	case <-outCh:
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("expected outbound message")
	}
}
