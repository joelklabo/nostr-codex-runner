# Extensions

This directory contains guidance and templates for adding new plugins.

## Transport template

- Create `internal/transports/<name>/<name>.go` implementing `core.Transport` (ID, Start, Send).
- Register with `transport.Register("<name>", ctor)` if you want dynamic loading.
- Use `internal/transports/mock` as a minimal reference; `internal/transports/nostr` shows a real one.

## Agent template

- Implement `core.Agent` in `internal/agents/<name>`, returning `core.AgentResponse` and optional `ActionCalls`.
- Register via `agent.Register("<name>", ctor)`.
- See `internal/agents/echo` (tiny) or `internal/agents/codexcli` (real CLI adapter).

## Action template

- Implement `core.Action` in `internal/actions/<name>`, declare capabilities, and register via `action.Register`.
- See `internal/actions/shell` and `internal/actions/fs`.

## Config mapping

- Extend `internal/config` structs to include your plugin-specific fields.
- Wire construction in `internal/app/build.go`.

See current implementations for working examples:

- Transports: `internal/transports/nostr`, `internal/transports/mock`, `internal/transports/slack` (stub)
- Agents: `internal/agents/codexcli`, `internal/agents/echo`, `internal/agents/http` (stub)
- Actions: `internal/actions/shell`, `internal/actions/fs`
