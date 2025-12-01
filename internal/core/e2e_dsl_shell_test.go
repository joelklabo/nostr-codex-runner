package core_test

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/joelklabo/buddy/internal/actions/shell"
	"github.com/joelklabo/buddy/internal/core"
	tmock "github.com/joelklabo/buddy/internal/transports/mock"
)

// DSL integration: /shell command triggers shell action and audit.
func TestRunnerDSLInvokeShellAction(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tr := tmock.New("mock")
	ag := &stubAgent{resp: core.AgentResponse{Reply: "noop"}}
	sh := shell.New(shell.Config{Allowed: []string{"echo "}, TimeoutSeconds: 5})

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	r := core.NewRunner([]core.Transport{tr}, ag, []core.Action{sh}, logger)

	done := make(chan struct{})
	go func() {
		_ = r.Start(ctx)
		close(done)
	}()

	tr.Inbound <- core.InboundMessage{Transport: "mock", Sender: "alice", Text: "/shell echo hi", ThreadID: "t2"}

	var out core.OutboundMessage
	select {
	case out = <-tr.Outbound:
	case <-time.After(3 * time.Second):
		t.Fatal("no outbound message")
	}

	cancel()
	<-done

	if !strings.Contains(out.Text, "hi") {
		t.Fatalf("expected shell output, got %q", out.Text)
	}
}
