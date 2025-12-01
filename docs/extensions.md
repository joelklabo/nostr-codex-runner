# Adding Extensions and Presets

This is for contributors adding new transports, agents, actions, or presets.

## CLI subcommands

- Commands live in `cmd/runner/main.go`. Keep the surface small; prefer presets/wizard over flags.

- Add a new case in `parseSubcommand` and implement in a small helper (`runXYZ`).

- Update `printHelp` and `docs/man/buddy.1.md` for consistency; add examples.

- Tests: `go test ./cmd/runner` and integration if applicable.

## Presets

- Add YAML under `internal/presets/data/<name>.yaml`.

- Register in `internal/presets/presets.go` (`//go:embed` + exported var) and `internal/presets/list.go`.

- Wizard: add to `internal/wizard/registry.go` presets list so it shows up in the picker.

- Dep hints: add to `PresetDeps()` in `internal/presets/presets.go` for `buddy check`/preflight.

- Docs: add a row in `docs/presets.md` and an example in README if prominent.

## Transports/agents/actions

- Add a new option to `internal/wizard/registry.go` with a short description.

- Implement transport/agent/action under `internal/transports|agents|actions`.

- If it needs deps, declare them in config `deps` map or `PresetDeps`.

- Add tests (table-driven, mock where possible) in the respective package.

## Commands vs. presets

- Prefer shipping a preset instead of a new subcommand. Subcommands are for core lifecycle (`run`, `wizard`, `presets`, `check`, `init-config`, `version`, `help`).

## Testing checklist

- Unit tests for new logic.

- `go test ./...`

- If wizard/presets affected, run `buddy wizard` dry-run and `buddy check <new-preset>`.

## Docs to update

- README (if user-facing).

- `docs/presets.md` (new preset).

- `docs/config.md` (new fields or defaults).

- `docs/wizard.md`/`docs/wizard-flow.md` if the flow changes.

- `docs/man/buddy.1.md` if CLI surface changes.
