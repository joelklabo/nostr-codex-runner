package nostrclient

import (
	"context"
	"testing"
	"time"

	"nostr-codex-runner/internal/store"

	"github.com/nbd-wtf/go-nostr"
	"sync"
)

func newStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.New(t.TempDir() + "/state.db")
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	return st
}

func TestIsReplay(t *testing.T) {
	c := &Client{lastMsg: make(map[string]lastSeen), windowDur: 2 * time.Second}
	ts := time.Now()
	if c.isReplay("a", "hi", ts) {
		t.Fatalf("first should not be replay")
	}
	if !c.isReplay("a", "hi", ts.Add(1*time.Second)) {
		t.Fatalf("second should be replay")
	}
	if c.isReplay("a", "hi", ts.Add(3*time.Second)) {
		t.Fatalf("window expired; should not be replay")
	}
}

func TestSenderLock(t *testing.T) {
	c := &Client{senderLocks: make(map[string]*sync.Mutex)}
	l1 := c.senderLock("a")
	l2 := c.senderLock("a")
	if l1 != l2 {
		t.Fatalf("expected same lock for same sender")
	}
	if c.senderLock("b") == l1 {
		t.Fatalf("different sender should get different lock")
	}
}

func TestAllowedList(t *testing.T) {
	st := newStore(t)
	c := New("k", "p", nil, []string{"a", "b"}, st)
	list := c.allowedList()
	if len(list) != 2 {
		t.Fatalf("expected 2 allowed, got %d", len(list))
	}
}

func TestLastCursorMax(t *testing.T) {
	st := newStore(t)
	defer st.Close()
	now := time.Now().UTC()
	_ = st.SaveCursor("a", now.Add(-10*time.Second))
	_ = st.SaveCursor("b", now.Add(-5*time.Second))
	c := &Client{
		store:   st,
		allowed: map[string]struct{}{"a": {}, "b": {}},
	}
	ts := c.lastCursorMax()
	if time.Unix(int64(ts), 0).After(now) {
		t.Fatalf("cursor should be <= now")
	}
}

func TestSharedSecretCaches(t *testing.T) {
	priv := nostr.GeneratePrivateKey()
	pub, _ := nostr.GetPublicKey(priv)
	c := New(priv, pub, nil, nil, newStore(t))
	sec1, err := c.sharedSecret(pub)
	if err != nil {
		t.Fatalf("shared secret: %v", err)
	}
	sec2, _ := c.sharedSecret(pub)
	if string(sec1) != string(sec2) {
		t.Fatalf("expected cached secret")
	}
}

func TestListenNilPool(t *testing.T) {
	c := &Client{pool: nil}
	if err := c.Listen(context.Background(), func(context.Context, IncomingMessage) {}); err == nil {
		t.Fatalf("expected error on nil pool")
	}
}
