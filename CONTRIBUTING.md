# Contributing

Thanks for helping improve the project! This repo uses `bd` for issue tracking and expects **one commit per closed issue**.

## How we work
- Current epic: `buddy-3oa` (buddy docs/wizard/CLI). Create or pick an issue under it: `bd create --parent buddy-3oa ...`.
- Keep work focused: one logical change per issue; close the issue with a commit that references it.
- Prefer small PRs with context and testing notes.
- Binary-first mindset: assume users install `buddy` via release/tap, not by cloning the repo.

## Development setup
- Requirements: Go ≥1.22, Codex CLI on PATH (for codex agent), `bd` installed, access to test Nostr relays.
- Install deps: `go mod download`
- Format: `gofmt -w ./cmd ./internal`
- Lint: `go vet ./...`
- Tests: `go test ./...`
- Run: `make run` (expects `config.yaml`) or `buddy run <preset|config>` once the binary rename lands.

## Commit and PR guidelines
- Reference the `bd` issue in the commit subject, e.g. `Quick install outline (closes buddy-3oa.2)`.
- Keep commit messages in imperative mood; avoid squashing unrelated work together.
- Include a short testing section in PR descriptions (commands run, notable results).

## Coding conventions
- Go style: idiomatic, small functions, explicit error wrapping, prefer `context.Context` plumbed.
- Config and secrets: never commit real keys; `.gitignore` already excludes `config.yaml` and state files.
- Logging: use structured logs (`slog`) and avoid leaking secret material.
- CLI naming: canonical binary will be `buddy`; ship `nostr-buddy` alias to avoid collisions. Note this in docs you touch.

## Security disclosures
See `SECURITY.md` for how to report vulnerabilities. Please **do not** open public issues for security findings.
