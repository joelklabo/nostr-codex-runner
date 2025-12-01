package store

import (
	"encoding/json"
	"testing"
	"time"
)

func TestHistoryAppendAndTrim(t *testing.T) {
	st, cleanup := newTempStore(t)
	defer cleanup()

	for i := 0; i < 5; i++ {
		b, _ := json.Marshal(i)
		if err := st.AppendHistory("thread", b, 3); err != nil {
			t.Fatalf("append err: %v", err)
		}
	}
	entries, err := st.History("thread", 10)
	if err != nil {
		t.Fatalf("history err: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
}

func TestAuditAppend(t *testing.T) {
	st, cleanup := newTempStore(t)
	defer cleanup()
	oldMax := auditMaxEntries
	auditMaxEntries = 2
	defer func() { auditMaxEntries = oldMax }()

	if err := st.AppendAudit("action", "sender", "ok", 0); err != nil {
		t.Fatalf("append audit: %v", err)
	}
	if err := st.AppendAudit("action", "sender", "ok2", 0); err != nil {
		t.Fatalf("append audit: %v", err)
	}
	if err := st.AppendAudit("action", "sender", "ok3", 0); err != nil {
		t.Fatalf("append audit: %v", err)
	}
	entries, err := st.Audit(10)
	if err != nil {
		t.Fatalf("audit err: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestAlreadyProcessedAndRecentMessage(t *testing.T) {
	st, cleanup := newTempStore(t)
	defer cleanup()

	if seen, err := st.AlreadyProcessed("e1"); err != nil || seen {
		t.Fatalf("first seen unexpected: %v %v", seen, err)
	}
	if seen, _ := st.AlreadyProcessed("e1"); !seen {
		t.Fatalf("second should be seen")
	}
	if err := st.MarkProcessed("e2"); err != nil {
		t.Fatalf("mark processed: %v", err)
	}

	seen, err := st.RecentMessageSeen("alice", "hello", time.Minute)
	if err != nil || seen {
		t.Fatalf("first recent should be false")
	}
	seen, _ = st.RecentMessageSeen("alice", "hello", time.Minute)
	if !seen {
		t.Fatalf("second recent should be true")
	}
}

func newTempStore(t *testing.T) (*Store, func()) {
	t.Helper()
	path := t.TempDir() + "/state.db"
	st, err := New(path)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	return st, func() { _ = st.Close() }
}

func TestCursorSaveAndLoad(t *testing.T) {
	st, cleanup := newTempStore(t)
	defer cleanup()

	now := time.Now().UTC().Truncate(time.Millisecond)
	if err := st.SaveCursor("alice", now); err != nil {
		t.Fatalf("save cursor: %v", err)
	}
	got, err := st.LastCursor("alice")
	if err != nil {
		t.Fatalf("last cursor: %v", err)
	}
	if !got.Equal(now) {
		t.Fatalf("cursor mismatch %v vs %v", got, now)
	}
}

func TestActiveLifecycle(t *testing.T) {
	st, cleanup := newTempStore(t)
	defer cleanup()

	if _, ok, err := st.Active("alice"); err != nil || ok {
		t.Fatalf("expected no active session")
	}
	if err := st.SaveActive("alice", "sess1"); err != nil {
		t.Fatalf("save active: %v", err)
	}
	if stt, ok, err := st.Active("alice"); err != nil || !ok || stt.SessionID != "sess1" {
		t.Fatalf("unexpected active %+v ok=%v err=%v", stt, ok, err)
	}
	if err := st.ClearActive("alice"); err != nil {
		t.Fatalf("clear active: %v", err)
	}
	if _, ok, _ := st.Active("alice"); ok {
		t.Fatalf("expected cleared session")
	}
}

func TestHistoryValidationAndProcessedErrors(t *testing.T) {
	st, cleanup := newTempStore(t)
	defer cleanup()

	if err := st.AppendHistory("", json.RawMessage(`{}`), 0); err == nil {
		t.Fatalf("expected thread id error")
	}
	if _, err := st.History("", 1); err == nil {
		t.Fatalf("expected history validation error")
	}
	if _, err := st.AlreadyProcessed(""); err == nil {
		t.Fatalf("expected error on empty id")
	}
	if err := st.MarkProcessed(""); err == nil {
		t.Fatalf("expected error on empty id")
	}
	if seen, err := st.RecentMessageSeen("bob", "hi", 0); err != nil || seen {
		t.Fatalf("recent message with default window should be false")
	}
}
