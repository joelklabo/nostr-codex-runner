# Extending buddy: commands, presets, transports, agents, actions – Issue 3oa.6

## Adding a CLI subcommand
- Location: `cmd/runner/main.go` routes to subcommands (will be renamed to buddy). Add a new command under `internal/commands` or similar when CLI refactor lands.
- Pattern: define a struct with `Run(ctx)` and register in the command router; keep flags minimal (prefer positional args).
- Tests: add unit tests for argument parsing and error codes.

## Adding a preset
- Add YAML under `assets/presets/<name>.yaml` (embedded at build time) with `meta` block (description, secrets, safety).
- Override path precedence: user `~/.config/buddy/presets/<name>.yaml` > embedded.
- Update `docs/presets.md` and ensure `buddy presets <name>` renders summary.
- Tests: golden files for preset listing and load/merge behavior.

## Adding a transport
- Create package `internal/transports/<name>` implementing `core.Transport` (`ID()`, `Start`, `Send`).
- Register in `internal/transports/registry.go` via `MustRegister("name", Constructor)`.
- Add config struct and hook into `internal/config` parsing.
- Tests: unit tests for config parsing and send/start behaviors; use mock relays where possible.

## Adding an agent
- Create `internal/agents/<name>` implementing `core.Agent` (`Generate`).
- Register in `internal/agents/registry.go`.
- Map config in `internal/config`; document fields in `docs/config.md` and add a preset if useful.
- Tests: table-driven tests for request/response mapping and error handling.

## Adding an action
- Create `internal/actions/<name>` implementing `core.Action` (`Invoke`).
- Register in `internal/actions/registry.go`.
- Declare config schema (allowlists, timeouts, limits). Add to docs.
- Tests: unit tests covering validation and execution limits.

## Security / logging checklist
- Never log secrets (keys, tokens, DM text). Use redaction helpers if added.
- Enforce allowlists where applicable (runner allowed_pubkeys, action roots, shell timeouts/max_output).
- Prefer context-aware timeouts on external calls.

## Docs updates per extension
- Add a short entry to plugin catalog (docs/plugins or presets doc).
- If user-facing, add a recipe/example and note prerequisites.

## Release/packaging
- Ensure new files are included in embedded assets if required (presets/config examples).
- If new deps are added, update `go.mod` and validate `goreleaser` build matrix.
