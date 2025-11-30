package http

import (
	"context"
	"errors"

	"nostr-codex-runner/internal/core"
)

// Config describes a generic HTTP LLM endpoint.
type Config struct {
	APIBase string
	Model   string
	APIKey  string
}

// Agent is a stub HTTP agent placeholder.
type Agent struct {
	cfg Config
}

func New(cfg Config) *Agent {
	return &Agent{cfg: cfg}
}

func (a *Agent) Generate(ctx context.Context, req core.AgentRequest) (core.AgentResponse, error) {
	return core.AgentResponse{}, errors.New("http agent not implemented")
}
