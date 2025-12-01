package core

import (
	"context"
	"encoding/json"
)

// Transport moves messages between an external system and the runner.
type Transport interface {
	// ID returns a stable identifier (e.g., "nostr-dm", "slack-webhook").
	ID() string
	// Start begins receiving inbound messages and pushing them into the provided channel.
	// It should return when ctx is canceled or a fatal error occurs.
	Start(ctx context.Context, inbound chan<- InboundMessage) error
	// Send delivers an outbound message back to the external system.
	Send(ctx context.Context, msg OutboundMessage) error
}

// Agent produces model-driven replies and optional action calls.
type Agent interface {
	Generate(ctx context.Context, req AgentRequest) (AgentResponse, error)
}

// Action exposes a callable capability (shell, fs, git, etc.).
type Action interface {
	Name() string
	Capabilities() []string
	// Help returns a short usage string (single line). Empty string means no help entry.
	Help() string
	Invoke(ctx context.Context, args json.RawMessage) (json.RawMessage, error)
}

// InboundMessage represents a message entering the runner.
type InboundMessage struct {
	Transport string         `json:"transport"`
	Sender    string         `json:"sender"`
	Text      string         `json:"text"`
	ThreadID  string         `json:"thread_id"`
	Meta      map[string]any `json:"meta,omitempty"`
}

// OutboundMessage represents a message leaving the runner.
type OutboundMessage struct {
	Transport string         `json:"transport"`
	Recipient string         `json:"recipient"`
	Text      string         `json:"text"`
	ThreadID  string         `json:"thread_id"`
	Meta      map[string]any `json:"meta,omitempty"`
}

// AgentRequest supplies the agent with prompt/context and available actions.
type AgentRequest struct {
	Prompt     string         `json:"prompt"`
	History    []MessageTurn  `json:"history,omitempty"`
	Actions    []ActionSpec   `json:"actions,omitempty"`
	SenderMeta map[string]any `json:"sender_meta,omitempty"`
}

// AgentResponse is produced by the agent.
type AgentResponse struct {
	Reply       string       `json:"reply"`
	SessionID   string       `json:"session_id,omitempty"`
	ActionCalls []ActionCall `json:"action_calls,omitempty"`
}

// MessageTurn represents one exchange in history.
type MessageTurn struct {
	Role string `json:"role"` // user or agent
	Text string `json:"text"`
}

// ActionSpec advertises an available action to the agent.
type ActionSpec struct {
	Name         string   `json:"name"`
	Capabilities []string `json:"capabilities,omitempty"`
	Description  string   `json:"description,omitempty"`
}

// ActionCall is an agent-requested invocation of an action.
type ActionCall struct {
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}
