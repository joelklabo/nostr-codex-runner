# Pluggable Agent Runner (Nostr example included)

[![CI](https://github.com/joelklabo/nostr-codex-runner/actions/workflows/ci.yml/badge.svg)](https://github.com/joelklabo/nostr-codex-runner/actions/workflows/ci.yml)
[![Release](https://github.com/joelklabo/nostr-codex-runner/actions/workflows/release.yml/badge.svg)](https://github.com/joelklabo/nostr-codex-runner/actions/workflows/release.yml)
[![Coverage](https://github.com/joelklabo/nostr-codex-runner/actions/workflows/coverage.yml/badge.svg)](https://github.com/joelklabo/nostr-codex-runner/actions/workflows/coverage.yml)
[![Staticcheck](https://github.com/joelklabo/nostr-codex-runner/actions/workflows/staticcheck.yml/badge.svg)](https://github.com/joelklabo/nostr-codex-runner/actions/workflows/staticcheck.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/joelklabo/nostr-codex-runner.svg)](https://pkg.go.dev/github.com/joelklabo/nostr-codex-runner)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](#license)
[![Latest Release](https://img.shields.io/github/v/release/joelklabo/nostr-codex-runner?sort=semver)](https://github.com/joelklabo/nostr-codex-runner/releases/latest)

Always-on bridge that listens for messages, feeds them into an AI agent, and executes optional host actions. Architecture is fully pluggable:
- **Transport**: how messages arrive/leave (built-ins: Nostr DM, mock; Slack stub scaffold).
- **Agent**: the model backend (built-ins: Codex CLI, echo, HTTP stub for OpenAI/Claude-style).
- **Action**: host capabilities (built-ins: shell exec, fs read/write; extend with your own).
The original Nostr Codex Runner is now just one config of this framework.

## Why
- Stay keyboard-only and remote: send prompts via Nostr DMs, get Codex replies back as DMs.
- Keep conversation context: runner tracks Codex `thread_id` per sender and resumes automatically.
- Minimal surface area: single binary, YAML config, and one background process.

## Command mini-DSL (DM payloads)
- `/new [prompt]` — reset session; optional prompt starts a fresh Codex session.
- `/use <session-id>` — switch to an existing Codex session.
- `/raw <cmd>` — execute a shell command on the host (working dir defaults to your home directory).
- `/status` — show your active session and last update time.
- `/help` — recap commands.
- _Anything else_ — treated as a prompt and executed in your active session (or a new one if none).

## Quick start (Nostr + Codex example)
1. Copy `config.example.yaml` → `config.yaml` and fill secrets:
   - `runner.private_key` — hex Nostr secret key (nsec).
   - `runner.allowed_pubkeys` — list of pubkeys allowed to control the runner.
   - Adjust relays, Codex working directory, timeouts, etc.
2. Run locally:
   ```bash
   make run              # uses config.yaml
   # or
   CONFIG=path/to/config.yaml ./scripts/run.sh
   ```
3. DM the runner pubkey from an allowed pubkey. Examples:
   ```
   /new
   /new Write a Go HTTP server that echoes requests.
   List the last 5 git commits in this repo.
   /status
   ```
4. Responses include `session: <thread-id>` plus the latest model message (truncated to `max_reply_chars`).

### Make targets
- `make run` – start with `config.yaml` (override `CONFIG=...`).
- `make build` – build to `bin/nostr-codex-runner` (override `BIN=...`).
- `make test` – run unit tests.
- `make lint` – `go vet ./...` (extra linters run in CI).
- `make fmt` – `gofmt -w cmd internal`.
- `make install` – `go install ./cmd/runner`.

## Install
- From source: `go install github.com/joelklabo/nostr-codex-runner/cmd/runner@latest`
- From release binaries (macOS/Linux amd64/arm64): grab the asset from the GitHub Releases page, `chmod +x nostr-codex-runner-*`, and run `./nostr-codex-runner --config config.yaml`.
- Docker image is not published yet; use the binary or source builds above.
- One-liner installer (downloads latest release to `~/.local/bin` and copies `config.example.yaml` → `config.yaml` if missing):
  ```bash
  curl -fsSL https://raw.githubusercontent.com/joelklabo/nostr-codex-runner/main/scripts/install.sh | bash
  ```
  Customize with env vars: `INSTALL_DIR`, `CONFIG_DIR`, `VERSION` (tag or `latest`).
- Prerequisite: Codex CLI must be installed and on `PATH` (full-access mode configured by default).
 - Full-access Codex: config sets `sandbox: danger-full-access`, `approval: never`, and `extra_args: ["--dangerously-bypass-approvals-and-sandbox"]` to give the agent unrestricted system access. Keep this only on trusted machines.

## Running it remotely / outside your LAN
The runner only needs outbound internet for its transport (e.g., Nostr relays). For shell access, rely on actions like `/raw` or your own VPN/Tailscale/SSH setup; there is no web UI. Optional health endpoint: run with `-health-listen 127.0.0.1:8081` for `/health` JSON.

## Quick links
- [Releases](https://github.com/joelklabo/nostr-codex-runner/releases)
- [Open issues](https://github.com/joelklabo/nostr-codex-runner/issues)
- [CI workflow](https://github.com/joelklabo/nostr-codex-runner/actions/workflows/ci.yml)
- [Release workflow](https://github.com/joelklabo/nostr-codex-runner/actions/workflows/release.yml)

## Configuration reference (plugins)
`config.example.yaml` documents every field. Key knobs:
- `transports[]: list of transports. Example (Nostr): see config.example.yaml; legacy relays still default to nostr.
- `runner.allowed_pubkeys`: access control.
- `runner.session_timeout_minutes`: idle cutoff before discarding a session mapping.
- `codex.*`: CLI flags for Codex (sandbox, approval policy, working dir (defaults to your home), extra args, timeout).
- `storage.path`: BoltDB file for state.
- `logging.level`: `debug|info|warn|error`; `logging.format`: `text|json`.

## Background service (macOS-friendly)
- tmux: `tmux new -s codex-runner 'cd /Users/honk/code/nostr-codex-runner && make run'`
- launchd (recommended for “always on”):
  - Make sure Codex and Node are on PATH (Homebrew defaults live in `/opt/homebrew/bin`). Either set `codex.binary` in `config.yaml` to an absolute path or pass PATH via the plist, e.g.:
  ```xml
  <?xml version="1.0" encoding="UTF-8"?>
  <!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
  <plist version="1.0">
  <dict>
    <key>Label</key><string>com.honk.nostr-codex-runner</string>
    <key>ProgramArguments</key>
    <array>
      <string>/Users/honk/bin/nostr-codex-runner</string>
      <string>-config</string>
      <string>/Users/honk/code/nostr-codex-runner/config.yaml</string>
    </array>
    <key>EnvironmentVariables</key>
    <dict>
      <key>PATH</key>
      <string>/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin</string>
    </dict>
    <key>WorkingDirectory</key><string>/Users/honk/code/nostr-codex-runner</string>
    <key>RunAtLoad</key><true/>
    <key>KeepAlive</key><true/>
    <key>StandardOutPath</key><string>/Users/honk/Library/Logs/nostr-codex-runner.log</string>
    <key>StandardErrorPath</key><string>/Users/honk/Library/Logs/nostr-codex-runner.err</string>
  </dict>
  </plist>
  ```
  Load/restart:
  ```
  launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/com.honk.nostr-codex-runner.plist
  launchctl kickstart -k gui/$(id -u)/com.honk.nostr-codex-runner
  ```

## Architecture (pluggable core)
- **Transports**: interface `Start/Send`; register via `internal/transports/registry.go`. Add new packages under `internal/transports/<name>`.
- **Agents**: interface `Generate`; Codex CLI is default, swap in HTTP/OpenAI/Claude via config.
- **Actions**: interface `Invoke`; shell/fs included. Register custom actions in `internal/actions`.
- **Runner**: wires transports→agent→actions, applies allowlists, mini-DSL, retry, session persistence, initial prompts, and max-reply truncation.
- **Config**: everything is declared in `config.yaml` (`transports[]`, `agent`, `actions[]`), so swapping Nostr→Slack or Codex→HTTP is config-only once the plugin exists.

## Development
- Requirements: Go ≥1.22, Codex CLI installed and on PATH for the Codex agent; other agents may have their own deps.
- Commands: `make run`, `make build`, `make test`, `make lint`, `make fmt`, `make install`.
- CI extras: coverage, staticcheck, misspell+gofmt, govulncheck, gosec, docker build, release.
- Formatting: `gofmt -w` on Go files before committing.
- Issue workflow: use `bd` (`bd create --parent nostr-codex-runner-2zo ...`) and close issues with one commit per issue.

## Security
- Keep `config.yaml` and keys private (already in `.gitignore`).
- Use trusted relays; consider a private relay for production.
- Limit `allowed_pubkeys` to operators you trust.
- Report vulnerabilities via a private GitHub security advisory (see `SECURITY.md`).

## Contributing
See `CONTRIBUTING.md` for how to propose changes, run checks, and follow the `bd`/commit conventions.

## License
MIT — see `LICENSE`.
- Initial prompt: `runner.initial_prompt` (prepended once for new sessions). Set it to remind the agent of its purpose.
