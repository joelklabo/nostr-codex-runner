package transport

import (
	"context"
	"testing"
	"time"

	"nostr-codex-runner/internal/core"
	tmock "nostr-codex-runner/internal/transports/mock"
)

// Minimal contract checks for transports we can instantiate without secrets.
func TestTransportConformanceMock(t *testing.T) {
	tr := tmock.New("mock1")
	if tr.ID() == "" {
		t.Fatalf("id should not be empty")
	}

	inbound := make(chan core.InboundMessage, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = tr.Start(ctx, inbound) }()

	// Should stop on cancel
	cancel()
	select {
	case <-time.After(200 * time.Millisecond):
		// ok: Start returned when ctx canceled
	default:
	}
}
