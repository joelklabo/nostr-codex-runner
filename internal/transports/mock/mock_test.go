package mock

import (
	"context"
	"testing"
	"time"

	"nostr-codex-runner/internal/core"
)

func TestMockTransportMovesMessages(t *testing.T) {
	tr := New("mock")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inbound := make(chan core.InboundMessage, 1)

	go func() {
		_ = tr.Start(ctx, inbound)
	}()

	tr.Inbound <- core.InboundMessage{Transport: "mock", Sender: "s", Text: "hi"}

	select {
	case msg := <-inbound:
		if msg.Text != "hi" {
			t.Fatalf("unexpected text: %s", msg.Text)
		}
	case <-time.After(time.Second):
		t.Fatal("no inbound received")
	}

	if err := tr.Send(ctx, core.OutboundMessage{Transport: "mock", Recipient: "s", Text: "ok"}); err != nil {
		t.Fatalf("send err: %v", err)
	}

	select {
	case out := <-tr.Outbound:
		if out.Text != "ok" {
			t.Fatalf("unexpected outbound: %s", out.Text)
		}
	case <-time.After(time.Second):
		t.Fatal("no outbound received")
	}
}
