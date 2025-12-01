# Plugin Architecture Overview

The runner is composed of three swappable parts:

- **Transport**: moves messages between an external system (Nostr, Slack, SMS, etc.) and the runner.
- **Agent**: calls an AI backend (Codex CLI, OpenAI/Claude, echo) to generate replies and optional action calls.
- **Action**: a callable capability on the host (shell, fs read/write, git, etc.), invoked by the agent.

Core data types (in `internal/core/types.go`):
- `InboundMessage{Transport, Sender, Text, ThreadID, Meta}`
- `OutboundMessage{Transport, Recipient, Text, ThreadID, Meta}`
- `AgentRequest{Prompt, History, Actions, SenderMeta}`
- `AgentResponse{Reply, SessionID, ActionCalls}`
- `ActionCall{Name, Args}`

Interfaces (in `internal/core/types.go`):
- `Transport{ID, Start(ctx,inbound), Send(ctx,msg)}`
- `Agent{Generate(ctx, AgentRequest) -> AgentResponse}`
- `Action{Name, Capabilities, Invoke(ctx,args)}`

Core runner (in `internal/core/runner.go`):
1. Transports push `InboundMessage` into the runner.
2. Runner parses command DSL (help/status/new/use/shell) and/or forwards prompt to Agent.
3. Agent may request `ActionCalls`; runner enforces allowlists and timeouts, executes actions, appends results.
4. Runner sends `OutboundMessage` via the same transport.

Registries:
- `internal/transportss/registry.go`
- `internal/agentss/registry.go`
- `internal/actionss/registry.go`

Built-in implementations:
- Transports: Nostr DM (`internal/transportss/nostr`), Mock (`internal/transportss/mock`), Slack stub (`internal/transportss/slack`).
- Agents: Codex CLI (`internal/agentss/codexcli`), Echo (`internal/agentss/echo`), HTTP stub (`internal/agentss/http`).
- Actions: Shell (`internal/actionss/shell`), FS read/write (`internal/actionss/fs`).

Configuration (see `config.example.yaml` and `configs/` samples):
- `transports[]`: `{type,id,...}` e.g., nostr/more fields per type.
- `agent`: `{type: codexcli|echo|http, codex: CodexConfig}`
- `actions[]`: `{type: shell|readfile|writefile, name, workdir/roots/etc.}`

State & policy:
- Storage interface in `internal/store/api.go`; Bolt implementation in `internal/store/store.go` with buckets for sessions, cursors, processed events, message replay, history, audit.
- Action policy: allowlist by name and sender; action timeouts; audit logging hooks.

Main wiring:
- `internal/app/build.go` maps config -> concrete transports/agent/actions and constructs `core.Runner`.
- `cmd/runner/main.go` loads config, opens Bolt store, builds runner, and starts it.

Extending:
- Add a transport: implement `core.Transport`, register constructor in a new package under `internal/transportss/<name>`, and extend config parsing.
- Add an agent: implement `core.Agent` and register; shape `AgentResponse.ActionCalls` to request tools.
- Add an action: implement `core.Action`, register, and add config schema; declare capabilities.
