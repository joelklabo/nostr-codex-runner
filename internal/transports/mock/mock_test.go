package mock

import (
	"context"
	"testing"

	"nostr-codex-runner/internal/core"
)

func TestMockStartForwardsInbound(t *testing.T) {
	tr := New("m1")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inbound := make(chan core.InboundMessage, 1)
	go func() { _ = tr.Start(ctx, inbound) }()

	msg := core.InboundMessage{Transport: "mock", Text: "hi"}
	tr.Inbound <- msg

	select {
	case got := <-inbound:
		if got.Text != "hi" {
			t.Fatalf("unexpected message %+v", got)
		}
	case <-ctx.Done():
		t.Fatalf("context canceled before message delivered")
	}
	cancel()
}

func TestMockSendRespectsContext(t *testing.T) {
	tr := New("m1")
	ctx := context.Background()
	outMsg := core.OutboundMessage{Text: "reply"}
	if err := tr.Send(ctx, outMsg); err != nil {
		t.Fatalf("send err: %v", err)
	}
	select {
	case got := <-tr.Outbound:
		if got.Text != "reply" {
			t.Fatalf("unexpected outbound %+v", got)
		}
	default:
		t.Fatalf("expected outbound message")
	}

	cctx, cancel := context.WithCancel(context.Background())
	cancel()
	// use unbuffered outbound to force ctx.Done path
	tr.Outbound = make(chan core.OutboundMessage)
	if err := tr.Send(cctx, outMsg); err == nil {
		t.Fatalf("expected error on canceled context")
	}
}
