# Repository Guidelines

## Project Structure & Module Organization

- `cmd/runner`: main entrypoint that wires config, Nostr client, command router, and Codex runner.
- `internal/commands`: mini-DSL for DM payloads (`/new`, `/use`, `/shell`, etc.).
- `internal/codex`: thin shell around `codex exec` and session tracking.
- `internal/config`: YAML parsing and defaults; copy `config.example.yaml` → `config.yaml` and edit there.
- `internal/nostrclient`: relay connections, DM decrypt/reply, replay protection.
- `internal/store`: BoltDB-backed state (default path in config). Keep test data out of source control.
- `scripts/`: helper shell scripts (`run.sh`, `install.sh`). Assets live in `assets/`.

## Build, Test, and Development Commands

- `make run` (optional `CONFIG=path`): start the runner using the given config.
- `make build` (optional `BIN=...`): build binary to `bin/nostr-codex-runner`.
- `make test` or `go test ./...`: run all Go tests.
- `make lint`: `go vet ./...`.
- `make fmt`: `gofmt -w cmd internal`.
- `go mod download`: fetch deps; run once after cloning.

## Coding Style & Naming Conventions

- Go ≥1.22; format with `gofmt` before committing. Keep functions small, plumb `context.Context`, wrap errors with detail.
- Logging: use structured `slog`; never log secrets (keys, tokens, decrypted DM text).
- Naming: package-level types and functions stay idiomatic Go (`CamelCase`); config keys match the YAML schema (lower_snake).

## Testing Guidelines

- Place tests beside code in `_test.go` files; prefer table-driven tests.
- Aim for coverage on parsing (commands/config) and replay protections; mock Nostr/Codex I/O where possible.
- Use `go test ./...` locally; add focused tests for regressions before merging.

## Commit & Pull Request Guidelines

- One issue per commit; reference the `bd` ticket in the subject, e.g. `a2d.2: harden DM dedup` or `Add replay debounce (closes nostr-codex-runner-2zo.3)`. Keep subjects imperative.
- PRs: include a short summary, linked issue, and **Testing** section listing commands run (`make test`, `make run` smoke, etc.). Note config/ops changes and any backward-incompatible defaults.
- Keep changes small and scoped; avoid combining refactors with feature work.

## Security & Configuration Tips

- Never commit real `config.yaml`, private keys, or state DBs; `.gitignore` already excludes them.
- Restrict `runner.allowed_pubkeys` to trusted operators and prefer private relays for production.
- If running with `sandbox: danger-full-access` or `approval: never`, do so only on trusted machines; otherwise tighten Codex CLI flags to limit blast radius.
