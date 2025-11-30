package codexcli

import (
	"context"
	"fmt"

	"nostr-codex-runner/internal/codex"
	"nostr-codex-runner/internal/config"
	"nostr-codex-runner/internal/core"
)

// Config mirrors config.CodexConfig for agent use.
type Config config.CodexConfig

// Agent wraps the Codex CLI runner.
type Agent struct {
	runner *codex.Runner
}

func New(cfg Config) *Agent {
	return &Agent{runner: codex.New(config.CodexConfig(cfg))}
}

func (a *Agent) Generate(ctx context.Context, req core.AgentRequest) (core.AgentResponse, error) {
	if req.Prompt == "" {
		return core.AgentResponse{}, fmt.Errorf("prompt is empty")
	}
	res, err := a.runner.Run(ctx, "", req.Prompt)
	if err != nil {
		return core.AgentResponse{}, err
	}
	return core.AgentResponse{Reply: res.Reply, SessionID: res.SessionID}, nil
}
