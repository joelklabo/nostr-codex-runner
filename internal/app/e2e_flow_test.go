package app

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/joelklabo/buddy/internal/assets"
	"github.com/joelklabo/buddy/internal/config"
	"github.com/joelklabo/buddy/internal/core"
	"github.com/joelklabo/buddy/internal/presets"
	"github.com/joelklabo/buddy/internal/store"
	tmock "github.com/joelklabo/buddy/internal/transports/mock"
)

// End-to-end smoke using mock transport + echo agent to ensure wiring works.
func TestE2EFlowWithMockTransport(t *testing.T) {
	cfg := &config.Config{
		Runner:     config.RunnerConfig{PrivateKey: "k", AllowedPubkeys: []string{"alice"}},
		Storage:    config.StorageConfig{Path: t.TempDir() + "/state.db"},
		Transports: []config.TransportConfig{{Type: "mock", ID: "mock"}},
		Agent:      config.AgentConfig{Type: "echo"},
		Actions:    []config.ActionConfig{{Type: "shell", Name: "shell", Workdir: "."}},
		Projects:   []config.Project{{ID: "p", Name: "p", Path: "."}},
	}
	st, err := store.New(cfg.Storage.Path)
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	defer func() { _ = st.Close() }()

	r, err := Build(cfg, st, slog.Default())
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	// Grab the mock transport to inject a message.
	var mt *tmock.Transport
	for _, tr := range r.Transports() {
		if candidate, ok := tr.(*tmock.Transport); ok {
			mt = candidate
			break
		}
	}
	if mt == nil {
		t.Fatalf("mock transport not found")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = r.Start(ctx) }()

	inbound := mt.Inbound
	outbound := mt.Outbound

	inbound <- core.InboundMessage{Transport: mt.ID(), Sender: "alice", Text: "hello", ThreadID: "t"}

	select {
	case msg := <-outbound:
		if msg.Text == "" {
			t.Fatalf("empty outbound")
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for outbound")
	}
}

// Ensures embedded mock-echo preset can run end-to-end once keys are filled.
func TestE2EFlowWithMockEchoPreset(t *testing.T) {
	td := t.TempDir()

	presetBytes, err := presets.Get("mock-echo")
	if err != nil {
		t.Fatalf("get preset: %v", err)
	}
	cfg, err := config.LoadBytes(presetBytes, td)
	if err != nil {
		t.Fatalf("load preset: %v", err)
	}
	// Fill required fields the preset leaves blank.
	cfg.Runner.PrivateKey = "mock"
	cfg.Runner.AllowedPubkeys = []string{"alice"}
	for i := range cfg.Transports {
		cfg.Transports[i].PrivateKey = "mock"
		cfg.Transports[i].AllowedPubkeys = []string{"alice"}
	}
	cfg.Storage.Path = td + "/state.db"
	cfg.Logging.File = "" // keep stdout only

	st, err := store.New(cfg.Storage.Path)
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	defer func() { _ = st.Close() }()

	r, err := Build(cfg, st, slog.Default())
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	var mt *tmock.Transport
	for _, tr := range r.Transports() {
		if candidate, ok := tr.(*tmock.Transport); ok {
			mt = candidate
			break
		}
	}
	if mt == nil {
		t.Fatalf("mock transport not found")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = r.Start(ctx) }()

	mt.Inbound <- core.InboundMessage{Transport: mt.ID(), Sender: "alice", Text: "hi", ThreadID: "t"}

	select {
	case msg := <-mt.Outbound:
		if msg.Text == "" {
			t.Fatalf("empty outbound")
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for outbound")
	}
	_ = assets.ConfigExample // keep embed reference reachable in test binary
}
