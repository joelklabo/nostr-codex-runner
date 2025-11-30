# Extensions

This directory contains guidance and templates for adding new plugins.

## Transport template
- Create `internal/transportss/<name>/<name>.go` implementing `core.Transport` (ID, Start, Send).
- Register with `transport.Register("<name>", ctor)` if you want dynamic loading.
- Follow Nostr and mock transports as references.

## Agent template
- Implement `core.Agent` in `internal/agentss/<name>`, returning `core.AgentResponse` and optional `ActionCalls`.
- Register via `agent.Register("<name>", ctor)`.

## Action template
- Implement `core.Action` in `internal/actionss/<name>`, declare capabilities, and register via `action.Register`.

## Config mapping
- Extend `internal/config` structs to include your plugin-specific fields.
- Wire construction in `internal/app/build.go`.

See current implementations for working examples:
- Transports: `internal/transportss/nostr`, `internal/transportss/mock`, `internal/transportss/slack` (stub)
- Agents: `internal/agentss/codexcli`, `internal/agentss/echo`, `internal/agentss/http` (stub)
- Actions: `internal/actionss/shell`, `internal/actionss/fs`
