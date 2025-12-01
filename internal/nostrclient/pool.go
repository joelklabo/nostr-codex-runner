package nostrclient

import (
	"context"

	"github.com/nbd-wtf/go-nostr"
)

// Pool abstracts SimplePool for testability.
type Pool interface {
	SubscribeMany(ctx context.Context, relays []string, filter nostr.Filter, opts ...nostr.SubscriptionOption) chan nostr.RelayEvent
	PublishMany(ctx context.Context, relays []string, ev nostr.Event) chan nostr.PublishResult
}
