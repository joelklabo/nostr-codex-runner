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
- **Copilot agent**: use GitHub Copilot CLI by setting `agent.type: copilotcli` (requires `copilot` from <https://github.com/github/copilot-cli>).
- **Action**: host capabilities (built-ins: shell exec, fs read/write; extend with your own).

The original Nostr Codex Runner is now just one config of this framework.

## Project layout (where plugins live)

- Transports: `internal/transports/<name>` (nostr, mock, slack stub)
- Agents: `internal/agents/<name>` (codexcli, copilotcli, echo, http stub)
- Actions: `internal/actions/<name>` (shell, fs read/write)
- Core runner/router: `internal/core`, `internal/commands`
- App wiring: `internal/app/build.go` (reads `config.yaml` and instantiates plugins)

See [Plugin catalog](docs/plugins/README.md) for the current list and how to add more.

## Supported plugins (shipped)

- **Transports:** `nostr`, `mock`, `slack` (stub scaffold), `whatsapp` (Twilio webhook + REST)
- **Agents:** `codexcli`, `copilotcli`, `echo`, `http` (stub)
- **Actions:** `shell`, `readfile`, `writefile`

These are all composable—pick any transport + one agent + any actions in `config.yaml`.

## Why

- Stay keyboard-only and remote: send prompts via Nostr DMs, get Codex replies back as DMs.
- Keep conversation context: runner tracks Codex `thread_id` per sender and resumes automatically.
- Minimal surface area: single binary, YAML config, and one background process.
- Inspiration: [warelay](https://github.com/steipete/warelay) for its plug-and-play relay approach.

## Command mini-DSL (DM payloads)

- `/new [prompt]` — reset session; optional prompt starts a fresh session.
- `/use <session-id>` — switch to an existing session.
- `/shell <cmd>` — execute a shell command (only if `shell` action is configured).
- `/status` — show your active session and last update time.
- `/help` — recap commands and list available actions.
- _Anything else_ — treated as a prompt and executed in your active session (or a new one if none).

## Quick start (Nostr + Codex example)

1. Copy `config.example.yaml` → `config.yaml` and fill secrets:
   - `runner.private_key` — hex Nostr secret key (nsec).
   - `runner.allowed_pubkeys` — list of pubkeys allowed to control the runner.
   - Adjust relays, Codex working directory, timeouts, etc.

2. Run locally:

   ```bash
   nostr-codex-runner run -config config.yaml   # or set NCR_CONFIG=path/to/config.yaml
   # make run still works if you prefer
   ```

3. DM the runner pubkey from an allowed pubkey. Examples:

   ```text
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
- From release binaries (macOS/Linux amd64/arm64): grab the asset from the GitHub Releases page, `chmod +x nostr-codex-runner-*`, and run `./nostr-codex-runner run -config config.yaml`.
- Docker image is not published yet; use the binary or source builds above.
- One-liner installer (downloads latest release to `~/.local/bin` and copies `config.example.yaml` → `config.yaml` if missing):

  ```bash
  curl -fsSL https://raw.githubusercontent.com/joelklabo/nostr-codex-runner/main/scripts/install.sh | bash
  ```

  Customize with env vars: `INSTALL_DIR`, `CONFIG_DIR`, `VERSION` (tag or `latest`).

- Prerequisites depend on the agent you choose:
  - Codex CLI: binary on `PATH`; optional full-access flags (`sandbox: danger-full-access`, `approval: never`, `extra_args: ["--dangerously-bypass-approvals-and-sandbox"]`) — only on trusted machines.
  - Copilot CLI: `npm install -g @github/copilot && copilot auth login`.
  - HTTP/echo agents: no extra deps beyond Go.

## Running it remotely / outside your LAN

The runner only needs outbound internet for its transport (e.g., Nostr relays). For shell access, rely on actions like `/shell` (if enabled) or your own VPN/Tailscale/SSH setup; there is no web UI. Optional health endpoint: run with `-health-listen 127.0.0.1:8081` for `/health` JSON.

## Quick links

- [Releases](https://github.com/joelklabo/nostr-codex-runner/releases)
- [Open issues](https://github.com/joelklabo/nostr-codex-runner/issues)
- [CI workflow](https://github.com/joelklabo/nostr-codex-runner/actions/workflows/ci.yml)
- [Release workflow](https://github.com/joelklabo/nostr-codex-runner/actions/workflows/release.yml)
- [Recipe: Nostr + Copilot CLI](docs/recipes/nostr-copilot-cli.md)
- [Recipe: WhatsApp + Codex](docs/recipes/whatsapp-codex.md)

## Configuration reference (plugins)

`config.example.yaml` documents every field. Key knobs:

- `transports[]`: list of transports. Example (Nostr): see config.example.yaml; legacy relays still default to nostr.
- `runner.allowed_pubkeys`: access control.
- `runner.session_timeout_minutes`: idle cutoff before discarding a session mapping.
- `runner.initial_prompt`: prepended once on a new session to set agent persona/guardrails.
- `agent.config.*`: CLI-style knobs for the selected agent (binary, working dir, extra args, timeout). `agent.codex` remains as a backward-compatible alias.
- `actions[]`: host capabilities; declare the ones you want (e.g., `shell`, `readfile`, `writefile`).
- `storage.path`: BoltDB file for state.
- `logging.level`: `debug|info|warn|error`; `logging.format`: `text|json`.

## Background service

- tmux: `tmux new -s codex-runner 'cd /Users/honk/code/nostr-codex-runner && make run'`
- launchd (recommended for "always on"): Create a plist file at `~/Library/LaunchAgents/com.honk.nostr-codex-runner.plist` with your configuration. Make sure Codex and Node are on PATH (Homebrew defaults live in `/opt/homebrew/bin`). Either set `codex.binary` in `config.yaml` to an absolute path or pass PATH via the plist `EnvironmentVariables` dict. Load/restart with:

  ```bash
  launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/com.honk.nostr-codex-runner.plist
  launchctl kickstart -k gui/$(id -u)/com.honk.nostr-codex-runner
  ```
- systemd (Linux): use the template at `scripts/systemd/nostr-codex-runner.service`. Example:

  ```bash
  # copy and edit the service to point at your user + config path
  sudo cp scripts/systemd/nostr-codex-runner.service /etc/systemd/system/nostr-codex-runner@youruser.service
  sudo systemctl daemon-reload
  sudo systemctl enable --now nostr-codex-runner@youruser.service
  ```

  The unit defaults to `ExecStart=%h/.local/bin/nostr-codex-runner run -config $NCR_CONFIG` and `NCR_CONFIG=%h/.config/nostr-codex-runner/config.yaml`; adjust if you install elsewhere.

## Architecture (pluggable core)

- **Transports**: interface `Start/Send`; register via `internal/transports/registry.go`. Add new packages under `internal/transports/<name>`.
- **Agents**: interface `Generate`; Codex CLI is default, swap in HTTP/OpenAI/Claude via config.
- **Actions**: interface `Invoke`; shell/fs included. Register custom actions in `internal/actions`.
- **Runner**: wires transports→agent→actions, applies allowlists, mini-DSL, retry, session persistence, initial prompts, and max-reply truncation.
- **Config**: everything is declared in `config.yaml` (`transports[]`, `agent`, `actions[]`), so swapping Nostr→Slack or Codex→HTTP is config-only once the plugin exists.

### Adding your own plugins

- Create a new folder under `internal/transports|agents|actions/<yourname>`, implement the interface, and call `registry.MustRegister` in `init()`.
- Add a config stanza referencing `type: "<yourname>"` and any custom fields you need.
- Extend `internal/app/build.go` to wire config → constructor.

### Example config: Nostr + Copilot CLI + basic actions

```yaml
transports:
  - type: nostr
    id: nostr
    relays: ["wss://relay.damus.io"]
    private_key: "<your_nsec_hex>"
    allowed_pubkeys: ["<operator_npub_hex>"]
agent:
  type: copilotcli
  config:               # generic agent config (codex remains as a legacy alias)
    binary: copilot     # install via: npm install -g @github/copilot
    working_dir: .
    timeout_seconds: 120
    extra_args: ["--allow-all-tools"]  # optional; lets Copilot apply edits/execute without prompts
actions:
  - type: shell
    name: shell
    workdir: .
    timeout_seconds: 30
    max_output: 4000
  - type: readfile
    roots: ["."]
    max_bytes: 65536
  - type: writefile
    roots: ["."]
    allow_write: true
    max_bytes: 65536
runner:
  allowed_pubkeys: ["<operator_npub_hex>"]
  max_reply_chars: 8000
  initial_prompt: "You are an agent running via Copilot CLI. Be concise and safe."
storage:
  path: ./state.db
logging:
  level: info
```

### Flow example: WhatsApp (Twilio) + Codex CLI

1. Configure `config.yaml` (or add alongside other transports):

   ```yaml
   transports:
     - type: whatsapp
       id: whatsapp
       config:
         account_sid: "ACxxxxxxxx"
         auth_token: "your_twilio_auth_token"
         from_number: "whatsapp:+15550001234"
         listen: ":8083"
         path: "/twilio/webhook"
         allowed_numbers: ["15555550100"]   # optional allowlist (E.164 without +)
   agent:
     type: codexcli
     config:
       binary: codex
       working_dir: .
       timeout_seconds: 900
   actions:
     - type: shell
     - type: readfile
     - type: writefile
   runner:
     max_reply_chars: 4000
     initial_prompt: "You are an agent responding to WhatsApp users. Be concise and safe."
   storage:
     path: ./state.db
   logging:
     level: info
   ```

2. Run `make run`.

3. Expose `listen` publicly (e.g., `ngrok http 8083`) and set your Twilio WhatsApp webhook URL to `https://<public>/twilio/webhook`.

4. Send a WhatsApp message from an allowed number; the runner replies via Codex.

### Slack stub config example (transport swap)

```yaml
transports:
  - type: slack
    id: slack
    token: "xoxb-..."          # to be implemented in slack transport
agent:
  type: echo
actions:
  - type: shell
projects:
  - id: default
    path: .
    name: default
```

## Development

- Requirements: Go ≥1.24.10, Codex CLI installed and on PATH for the Codex agent; other agents may have their own deps.
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
