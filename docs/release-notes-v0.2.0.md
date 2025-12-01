# Release notes: v0.2.0 (plugin architecture)

Highlights
- Introduced pluggable interfaces (Transport, Agent, Action) and core runner orchestration.
- Added registries and adapters: Nostr DM, mock transport, Slack stub; Codex CLI, echo agent, HTTP stub; shell and fs read/write actions.
- Configuration extended with `transports`, `agent`, `actions` while keeping legacy keys as defaults.
- Action policy hooks: allowlist, sender checks, timeouts, audit logging; history/audit buckets added to Bolt store.
- Main rewired through `internal/app.Build`, cleaning direct Nostr/Codex coupling.

Samples & docs
- Sample configs (now renamed): `configs/copilot-shell.yaml`, `configs/mock-echo.yaml`.
- Architecture doc: `docs/architecture.md`.
- Migration guide: `docs/migration-v0.2.0.md`.

Notes
- Slack and HTTP agents are stubs; fill in real implementations as needed.
- Existing state DBs remain valid; new buckets are added for history/audit.
