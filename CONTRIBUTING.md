# Contributing

Thanks for helping improve the project! This repo uses `bd` for issue tracking and expects **one commit per closed issue**.

## How we work
- Current epic: `buddy-3oa` (buddy docs/wizard/CLI). Create or pick an issue under it: `bd create --parent buddy-3oa ...`.
- Keep work focused: one logical change per issue; close the issue with a commit that references it.
- Prefer small PRs with context and testing notes.
- Binary-first mindset: assume users install `buddy` via release/tap, not by cloning the repo.

## Development setup
- Requirements: Go ≥1.24, `bd` installed, access to test Nostr relays. Copilot CLI or HTTP keys are optional, depending on what you’re working on.
- Install deps: `go mod download`
- Format: `gofmt -w ./cmd ./internal`
- Lint: `go vet ./...`
- Tests: `go test ./...`
- Run: `make run` (expects `config.yaml`) or `buddy run <preset|config>` with embedded presets.

## Commit and PR guidelines
- Reference the `bd` issue in the commit subject, e.g. `Quick install outline (closes buddy-3oa.2)`.
- Keep commit messages in imperative mood; avoid squashing unrelated work together.
- Include a short testing section in PR descriptions (commands run, notable results).

## Coding conventions
- Go style: idiomatic, small functions, explicit error wrapping, prefer `context.Context` plumbed.
- Config and secrets: never commit real keys; `.gitignore` already excludes `config.yaml` and state files.
- Logging: use structured logs (`slog`) and avoid leaking secret material.
- CLI naming: canonical binary is `buddy`; we do not ship aliases. If you need `bud`, create a local alias and mention collisions in docs if relevant.
- Wizards/presets: prefer adding a preset before inventing flags; update `internal/presets`, `internal/wizard/registry`, and docs as part of the change.

## Security disclosures
See `SECURITY.md` for how to report vulnerabilities. Please **do not** open public issues for security findings.
