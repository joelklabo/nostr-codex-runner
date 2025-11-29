# Nostr Codex Runner

[![CI](https://github.com/joelklabo/nostr-codex-runner/actions/workflows/ci.yml/badge.svg)](https://github.com/joelklabo/nostr-codex-runner/actions/workflows/ci.yml)
[![Release](https://github.com/joelklabo/nostr-codex-runner/actions/workflows/release.yml/badge.svg)](https://github.com/joelklabo/nostr-codex-runner/actions/workflows/release.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/joelklabo/nostr-codex-runner.svg)](https://pkg.go.dev/github.com/joelklabo/nostr-codex-runner)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](#license)

Always-on bridge that listens for Nostr encrypted DMs from trusted pubkeys and feeds them into `codex exec`, keeping Codex session threads alive so you can work entirely over Nostr.

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

## Quick start
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
4. Responses include `session: <thread-id>` plus the latest Codex message (truncated to `max_reply_chars`).

## Install
- From source: `go install github.com/joelklabo/nostr-codex-runner/cmd/runner@latest`
- From release binaries (macOS/Linux amd64/arm64): grab the asset from the GitHub Releases page, `chmod +x nostr-codex-runner-*`, and run `./nostr-codex-runner --config config.yaml`.
- Docker image is not published yet; use the binary or source builds above.

## Quick links
- [Releases](https://github.com/joelklabo/nostr-codex-runner/releases)
- [Open issues](https://github.com/joelklabo/nostr-codex-runner/issues)
- [CI workflow](https://github.com/joelklabo/nostr-codex-runner/actions/workflows/ci.yml)
- [Release workflow](https://github.com/joelklabo/nostr-codex-runner/actions/workflows/release.yml)

### Web UI (local)
- Default: enabled at `http://127.0.0.1:8080`.
- Create epics or issues, edit existing items, and pick which project they belong to via a dropdown.
- Labels: add labels on create; add/remove labels on edit.
- Filtering: issue list can be filtered by status from the UI.
- Auth: set `ui.auth_token` in `config.yaml` to require `Authorization: Bearer <token>` for all UI/API calls (recommended if you expose the port beyond localhost).
- Projects come from `projects` in `config.yaml` (each `path` should contain the `.beads` directory for that repo).

## Configuration reference
`config.example.yaml` documents every field. Key knobs:
- `relays`: list of relay URLs to connect to.
- `runner.allowed_pubkeys`: access control.
- `runner.session_timeout_minutes`: idle cutoff before discarding a session mapping.
- `codex.*`: CLI flags for Codex (sandbox, approval policy, working dir (defaults to your home), extra args, timeout).
- `ui.*`: toggle/address for the local UI, optional `auth_token`; update `projects` to expose multiple bd workspaces in the dropdown.
- `storage.path`: BoltDB file for state.
- `logging.level`: `debug|info|warn|error`.

## Background service (macOS-friendly)
- tmux: `tmux new -s codex-runner 'cd /Users/honk/code/nostr-codex-runner && make run'`
- launchd: create `~/Library/LaunchAgents/com.honk.nostr-codex-runner.plist`:
  ```xml
  <?xml version="1.0" encoding="UTF-8"?>
  <!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
  <plist version="1.0">
  <dict>
    <key>Label</key><string>com.honk.nostr-codex-runner</string>
    <key>ProgramArguments</key>
    <array>
      <string>/usr/bin/env</string>
      <string>bash</string>
      <string>-lc</string>
      <string>cd /Users/honk/code/nostr-codex-runner && CONFIG=/Users/honk/code/nostr-codex-runner/config.yaml make run</string>
    </array>
    <key>RunAtLoad</key><true/>
    <key>KeepAlive</key><true/>
    <key>StandardOutPath</key><string>/Users/honk/Library/Logs/nostr-codex-runner.log</string>
    <key>StandardErrorPath</key><string>/Users/honk/Library/Logs/nostr-codex-runner.err</string>
  </dict>
  </plist>
  ```
  Load with `launchctl load ~/Library/LaunchAgents/com.honk.nostr-codex-runner.plist`.

## Architecture (short)
- **Nostr client**: subscribes to kind-4 DMs from allowlisted authors to runner pubkey; decrypts via NIP-04; deduplicates per-event ID.
- **Command router**: parses the mini-DSL; manages per-sender active session stored in BoltDB with idle expiry.
- **Codex runner**: shells out to `codex exec --json`, captures `thread_id` and latest `agent_message`.
- **Reply**: sends encrypted DM back with session id and truncated message.

## Development
- Requirements: Go ≥1.22, Codex CLI installed and on PATH.
- Commands:
  - `make run` — start the service (uses `config.yaml`).
  - `make build` — build binary to `bin/nostr-codex-runner`.
  - `make lint` — `go vet ./...`.
  - `go test ./...` — run tests (currently none; add as you extend).
- Formatting: `gofmt -w` on Go files before committing.
- Issue workflow: use `bd` (`bd create --parent nostr-codex-runner-2zo ...`) and close issues with a commit per issue.

## Security
- Keep `config.yaml` and keys private (already in `.gitignore`).
- Use trusted relays; consider a private relay for production.
- Limit `allowed_pubkeys` to operators you trust.
- Report vulnerabilities via a private GitHub security advisory (see `SECURITY.md`).

## Contributing
See `CONTRIBUTING.md` for how to propose changes, run checks, and follow the `bd`/commit conventions.

## License
MIT — see `LICENSE`.
