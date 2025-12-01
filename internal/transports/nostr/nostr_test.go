package nostr

import (
	"context"
	"errors"
	"testing"

	"nostr-codex-runner/internal/core"
	client "nostr-codex-runner/internal/nostrclient"
	"nostr-codex-runner/internal/store"

	"github.com/nbd-wtf/go-nostr"
)

func TestNewMissingKey(t *testing.T) {
	if _, err := New(Config{}, nil); err == nil {
		t.Fatalf("expected error for missing key")
	}
}

func TestNewAndID(t *testing.T) {
	priv := nostr.GeneratePrivateKey()
	st, _ := store.New(t.TempDir() + "/state.db")
	defer func() { _ = st.Close() }()
	tr, err := New(Config{PrivateKey: priv}, st)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if tr.ID() != "nostr" {
		t.Fatalf("id mismatch")
	}
}

type stubClient struct {
	listenErr error
	sendErr   error
	called    bool
}

func (s *stubClient) Listen(ctx context.Context, handler func(context.Context, client.IncomingMessage)) error {
	s.called = true
	return s.listenErr
}

func (s *stubClient) SendReply(ctx context.Context, toPubKey string, message string) error {
	return s.sendErr
}

func TestSendMissingRecipient(t *testing.T) {
	priv := nostr.GeneratePrivateKey()
	st, _ := store.New(t.TempDir() + "/state.db")
	defer func() { _ = st.Close() }()
	tr, _ := New(Config{PrivateKey: priv}, st)
	tr.client = &stubClient{}

	if err := tr.Send(context.Background(), core.OutboundMessage{}); err == nil {
		t.Fatalf("expected error for missing recipient")
	}
}

func TestSendPropagatesClientError(t *testing.T) {
	priv := nostr.GeneratePrivateKey()
	st, _ := store.New(t.TempDir() + "/state.db")
	defer func() { _ = st.Close() }()
	tr, _ := New(Config{PrivateKey: priv}, st)
	sc := &stubClient{sendErr: errors.New("boom")}
	tr.client = sc

	if err := tr.Send(context.Background(), core.OutboundMessage{Recipient: "npub1", Text: "hi"}); err == nil {
		t.Fatalf("expected send error to bubble")
	}
}

func TestStartReturnsClientError(t *testing.T) {
	priv := nostr.GeneratePrivateKey()
	st, _ := store.New(t.TempDir() + "/state.db")
	defer func() { _ = st.Close() }()
	tr, _ := New(Config{PrivateKey: priv}, st)
	sc := &stubClient{listenErr: errors.New("listen fail")}
	tr.client = sc

	err := tr.Start(context.Background(), make(chan core.InboundMessage))
	if err == nil || err.Error() != "listen fail" {
		t.Fatalf("unexpected err: %v", err)
	}
}
