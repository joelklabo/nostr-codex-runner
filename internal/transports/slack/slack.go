package slack

import (
	"context"
	"errors"
	"nostr-codex-runner/internal/core"
)

// Config for Slack transport (stub).
type Config struct {
	BotToken        string
	SigningSecret   string
	AllowedChannels []string
}

// Transport is a stub Slack transport implementing core.Transport.
type Transport struct {
	id  string
	cfg Config
}

func New(cfg Config) *Transport {
	return &Transport{id: "slack", cfg: cfg}
}

func (t *Transport) ID() string { return t.id }

func (t *Transport) Start(ctx context.Context, inbound chan<- core.InboundMessage) error {
	// Not implemented yet; placeholder for Events API/webhooks.
	<-ctx.Done()
	return ctx.Err()
}

func (t *Transport) Send(ctx context.Context, msg core.OutboundMessage) error {
	return errors.New("slack transport not implemented")
}
