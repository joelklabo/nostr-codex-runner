# Architecture Overview

Text-first diagram of the main flow:

1. Transport (Nostr/mock) receives a message/DM.

2. Router forwards to Agent (copilotcli/http/codexcli/echo).

3. Agent may call Actions (shell/readfile/writefile) via runner.

4. Responses go back through the same Transport.

Components:

- `internal/transports/*` — adapters (nostr, mock).

- `internal/agents/*` — LLM backends.

- `internal/actions/*` — host capabilities.

- `internal/app` — wires transport → agent → actions.

- `internal/presets` — embedded configs; `internal/wizard` — guided config.

- `cmd/runner` — CLI entry (`buddy run`, `wizard`, `presets`, `check`).

Data paths:

- Config (`config.yaml` or preset) → `internal/config`.

- State DB (BoltDB) → `internal/store`.

- Metrics/health → `internal/metrics`, `internal/health`.

Lifecycle:

- Start `buddy run <preset|config>`.

- Config loaded, dep preflight runs, transport connects, agent initialized.

- Each DM/session keeps context in memory + BoltDB cursors.

Notes:

- Dependency checks run before start (`buddy check` or run preflight).

- Offline default: `mock-echo` preset.
