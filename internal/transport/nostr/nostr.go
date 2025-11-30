package nostr

import (
	"context"
	"fmt"

	"nostr-codex-runner/internal/core"
	client "nostr-codex-runner/internal/nostrclient"
	"nostr-codex-runner/internal/store"

	"github.com/nbd-wtf/go-nostr"
)

// Config holds the parameters needed to run the Nostr transport.
type Config struct {
	Relays         []string
	PrivateKey     string
	AllowedPubkeys []string
}

// Transport implements core.Transport for Nostr DMs.
type Transport struct {
	cfg    Config
	store  *store.Store
	client *client.Client
	id     string
}

// New creates a Nostr transport.
func New(cfg Config, st *store.Store) (*Transport, error) {
	if cfg.PrivateKey == "" {
		return nil, fmt.Errorf("nostr private key required")
	}
	pub, err := nostr.GetPublicKey(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("derive pubkey: %w", err)
	}
	c := client.New(cfg.PrivateKey, pub, cfg.Relays, cfg.AllowedPubkeys, st)
	return &Transport{cfg: cfg, store: st, client: c, id: "nostr"}, nil
}

// ID returns transport identifier.
func (t *Transport) ID() string { return t.id }

// Start subscribes to Nostr DMs and pushes inbound messages.
func (t *Transport) Start(ctx context.Context, inbound chan<- core.InboundMessage) error {
	handler := func(msgCtx context.Context, msg client.IncomingMessage) {
		inbound <- core.InboundMessage{
			Transport: t.id,
			Sender:    msg.SenderPubKey,
			Text:      msg.Plaintext,
			ThreadID:  msg.SenderPubKey,
		}
	}
	return t.client.Listen(ctx, handler)
}

// Send delivers a DM reply back to sender.
func (t *Transport) Send(ctx context.Context, msg core.OutboundMessage) error {
	if msg.Recipient == "" {
		return fmt.Errorf("nostr recipient missing")
	}
	return t.client.SendReply(ctx, msg.Recipient, msg.Text)
}

func init() {
	// Allow registry registration when imported via blank import.
}
