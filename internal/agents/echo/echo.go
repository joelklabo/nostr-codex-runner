package echo

import (
	"context"

	"nostr-codex-runner/internal/core"
)

// Agent echoes prompts; intended for tests and examples.
type Agent struct{}

func New() *Agent { return &Agent{} }

func (a *Agent) Generate(ctx context.Context, req core.AgentRequest) (core.AgentResponse, error) {
	reply := req.Prompt
	return core.AgentResponse{Reply: reply}, nil
}
