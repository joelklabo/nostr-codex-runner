package nostrclient

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"nostr-codex-runner/internal/store"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
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
	defer func() { _ = st.Close() }()
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

// stub implementations for Listen testing
type stubStore struct {
	processed map[string]bool
}

func (s *stubStore) SaveActive(pubkey, sessionID string) error { return nil }
func (s *stubStore) ClearActive(pubkey string) error           { return nil }
func (s *stubStore) Active(pubkey string) (store.SessionState, bool, error) {
	return store.SessionState{}, false, nil
}
func (s *stubStore) LastCursor(pubkey string) (time.Time, error)  { return time.Time{}, nil }
func (s *stubStore) SaveCursor(pubkey string, ts time.Time) error { return nil }
func (s *stubStore) AlreadyProcessed(eventID string) (bool, error) {
	if s.processed == nil {
		s.processed = map[string]bool{}
	}
	seen := s.processed[eventID]
	s.processed[eventID] = true
	return seen, nil
}
func (s *stubStore) MarkProcessed(eventID string) error { return nil }
func (s *stubStore) RecentMessageSeen(pubkey, message string, window time.Duration) (bool, error) {
	return false, nil
}

type stubPool struct {
	ch chan nostr.RelayEvent
}

func newStubPool() *stubPool { return &stubPool{ch: make(chan nostr.RelayEvent, 1)} }

func (p *stubPool) SubscribeMany(ctx context.Context, relays []string, filter nostr.Filter, _ ...nostr.SubscriptionOption) chan nostr.RelayEvent {
	return p.ch
}
func (p *stubPool) PublishMany(ctx context.Context, relays []string, ev nostr.Event) chan nostr.PublishResult {
	out := make(chan nostr.PublishResult, 1)
	close(out)
	return out
}

type errPool struct{}

func (p errPool) SubscribeMany(ctx context.Context, relays []string, filter nostr.Filter, _ ...nostr.SubscriptionOption) chan nostr.RelayEvent {
	ch := make(chan nostr.RelayEvent)
	close(ch)
	return ch
}

func (p errPool) PublishMany(ctx context.Context, relays []string, ev nostr.Event) chan nostr.PublishResult {
	ch := make(chan nostr.PublishResult, 1)
	ch <- nostr.PublishResult{Error: errors.New("publish boom")}
	close(ch)
	return ch
}

type okPool struct{}

func (p okPool) SubscribeMany(ctx context.Context, relays []string, filter nostr.Filter, _ ...nostr.SubscriptionOption) chan nostr.RelayEvent {
	ch := make(chan nostr.RelayEvent)
	close(ch)
	return ch
}
func (p okPool) PublishMany(ctx context.Context, relays []string, ev nostr.Event) chan nostr.PublishResult {
	ch := make(chan nostr.PublishResult, 1)
	ch <- nostr.PublishResult{}
	close(ch)
	return ch
}

func TestListenProcessesInbound(t *testing.T) {
	priv := nostr.GeneratePrivateKey()
	pub, _ := nostr.GetPublicKey(priv)
	pool := newStubPool()
	st := &stubStore{}
	c := NewWithPool(priv, pub, []string{"wss://relay"}, []string{pub}, st, pool)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	got := make(chan IncomingMessage, 1)
	go func() { _ = c.Listen(ctx, func(_ context.Context, m IncomingMessage) { got <- m; cancel() }) }()

	ev := &nostr.Event{ID: "e1", PubKey: pub, CreatedAt: nostr.Now(), Kind: nostr.KindEncryptedDirectMessage, Tags: nostr.Tags{nostr.Tag{"p", pub}}}
	secret, _ := nip04.ComputeSharedSecret(pub, priv)
	enc, _ := nip04.Encrypt("hello", secret)
	ev.Content = enc
	if err := ev.Sign(priv); err != nil {
		t.Fatalf("sign: %v", err)
	}
	pool.ch <- nostr.RelayEvent{Event: ev}

	select {
	case msg := <-got:
		if msg.Plaintext != "hello" {
			t.Fatalf("plaintext mismatch: %s", msg.Plaintext)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout")
	}
}

func TestSendReplyPublishesErrors(t *testing.T) {
	priv := nostr.GeneratePrivateKey()
	pub, _ := nostr.GetPublicKey(priv)
	c := NewWithPool(priv, pub, []string{"wss://relay"}, []string{pub}, newStore(t), errPool{})

	err := c.SendReply(context.Background(), pub, "hi")
	if err == nil || err.Error() != "publish boom" {
		t.Fatalf("expected publish error, got %v", err)
	}
}

func TestPublishProfileSkipsEmpty(t *testing.T) {
	priv := nostr.GeneratePrivateKey()
	pub, _ := nostr.GetPublicKey(priv)
	c := NewWithPool(priv, pub, nil, nil, newStore(t), errPool{})
	if err := c.PublishProfile(context.Background(), "", ""); err != nil {
		t.Fatalf("expected nil with empty meta, got %v", err)
	}
}

func TestPublishProfilePropagatesError(t *testing.T) {
	priv := nostr.GeneratePrivateKey()
	pub, _ := nostr.GetPublicKey(priv)
	c := NewWithPool(priv, pub, []string{"wss://relay"}, []string{pub}, newStore(t), errPool{})
	if err := c.PublishProfile(context.Background(), "runner", "pic"); err == nil {
		t.Fatalf("expected publish error")
	}
}

func TestSendReplySuccess(t *testing.T) {
	priv := nostr.GeneratePrivateKey()
	pub, _ := nostr.GetPublicKey(priv)
	c := NewWithPool(priv, pub, []string{"wss://relay"}, []string{pub}, newStore(t), okPool{})
	if err := c.SendReply(context.Background(), pub, "hi"); err != nil {
		t.Fatalf("send reply: %v", err)
	}
}

func TestBuildFilterUsesAllowedAndCursor(t *testing.T) {
	st := newStore(t)
	defer func() { _ = st.Close() }()
	now := time.Now()
	_ = st.SaveCursor("alice", now.Add(-5*time.Second))
	priv := nostr.GeneratePrivateKey()
	pub, _ := nostr.GetPublicKey(priv)
	c := New(priv, pub, []string{"wss://relay"}, []string{"alice"}, st)
	f := c.buildFilter()
	if len(f.Authors) != 1 || f.Authors[0] != "alice" {
		t.Fatalf("authors mismatch: %+v", f.Authors)
	}
	if f.Since == nil {
		t.Fatalf("since missing")
	}
}

func TestPublishProfileSuccess(t *testing.T) {
	priv := nostr.GeneratePrivateKey()
	pub, _ := nostr.GetPublicKey(priv)
	c := NewWithPool(priv, pub, []string{"wss://relay"}, []string{pub}, newStore(t), okPool{})
	if err := c.PublishProfile(context.Background(), "runner", "pic"); err != nil {
		t.Fatalf("publish profile: %v", err)
	}
}
