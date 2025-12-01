package nostr

import (
	"context"
	"errors"
	"testing"

	"nostr-codex-runner/internal/core"
	"nostr-codex-runner/internal/nostrclient"
	"nostr-codex-runner/internal/store"

	"github.com/nbd-wtf/go-nostr"
)

type poolFail struct{}

func (p *poolFail) SubscribeMany(ctx context.Context, relays []string, filter nostr.Filter, _ ...nostr.SubscriptionOption) chan nostr.RelayEvent {
	ch := make(chan nostr.RelayEvent)
	close(ch)
	return ch
}
func (p *poolFail) PublishMany(ctx context.Context, relays []string, ev nostr.Event) chan nostr.PublishResult {
	ch := make(chan nostr.PublishResult, 1)
	ch <- nostr.PublishResult{Error: errors.New("fail")}
	close(ch)
	return ch
}

func TestSendErrorPropagation(t *testing.T) {
	priv := nostr.GeneratePrivateKey()
	pub, _ := nostr.GetPublicKey(priv)
	st, _ := store.New(t.TempDir() + "/state.db")
	defer func() { _ = st.Close() }()
	client := nostrclient.NewWithPool(priv, pub, nil, nil, st, &poolFail{})
	tr := &Transport{cfg: Config{PrivateKey: priv}, store: st, client: client, id: "nostr"}
	err := tr.Send(context.Background(), core.OutboundMessage{Recipient: "bob", Text: "hi"})
	if err == nil {
		t.Fatalf("expected send error")
	}
}
