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
