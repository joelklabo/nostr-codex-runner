package nostrclient

import (
	"context"
	"errors"

	"github.com/nbd-wtf/go-nostr"
)

// simplePoolStub lets us test Listen/Send without network.
type simplePoolStub struct {
	subCh  chan nostr.RelayEvent
	pubErr error
	subErr error
	closed bool
}

func newPoolStub() *simplePoolStub {
	return &simplePoolStub{subCh: make(chan nostr.RelayEvent, 1)}
}

func (s *simplePoolStub) SubscribeMany(ctx context.Context, _ []string, _ nostr.Filter, _ ...nostr.SubscriptionOption) chan nostr.RelayEvent {
	if s.subErr != nil {
		close(s.subCh)
	}
	return s.subCh
}

func (s *simplePoolStub) PublishMany(ctx context.Context, _ []string, _ nostr.Event) chan nostr.PublishResult {
	ch := make(chan nostr.PublishResult, 1)
	if s.pubErr != nil {
		ch <- nostr.PublishResult{Error: s.pubErr}
	}
	close(ch)
	return ch
}

// helper for tests to inject an event
func (s *simplePoolStub) emit(e *nostr.Event) {
	s.subCh <- nostr.RelayEvent{Event: e}
}

var errStub = errors.New("stub")
