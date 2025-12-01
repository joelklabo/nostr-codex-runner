# Architecture Overview (buddy CLI) – includes diagram for issue 3oa.5

```mermaid
flowchart LR
    subgraph CLI["buddy CLI"]
      A1[b u d d y<br/>run / wizard / presets]
      A2[Preset loader<br/>embedded + user overrides]
    end
    subgraph Config["Config & Presets"]
      C1[config.yaml<br/>(argv / cwd / ~/.config/buddy)]
      C2[Presets<br/>(embedded, ~/.config/buddy/presets, ./presets)]
    end
    subgraph Runner["Core runner"]
      R1[Command DSL<br/>/new /use /shell /status]
      R2[Session manager<br/>state in BoltDB]
      R3[Action executor<br/>allowlist + timeouts]
    end
    subgraph Plumbing["Plugins"]
      T[Transport(s)<br/>nostr/mock/slack/whatsapp]
      G[Agent<br/>codexcli/copilotcli/http/echo]
      X[Actions<br/>shell/readfile/writefile]
    end
    subgraph Observability["Observability"]
      M[Metrics /health / Prom]
      L[Structured logs]
    end
    A1 --> A2 --> C2
    A1 --> C1
    C1 --> R1
    C2 --> R1
    R1 --> R2 --> G
    G --> R3
    R3 --> X
    R1 --> T
    R2 --> T
    R2 --> M
    R2 --> L
```

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
