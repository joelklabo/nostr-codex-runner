# Recipe: Nostr transport + GitHub Copilot CLI agent

This walks through running the pluggable runner with:
- **Transport**: Nostr DMs.
- **Agent**: GitHub Copilot CLI (`gh copilot suggest`).
- **Actions**: shell, readfile, and writefile.

## Prerequisites
- GitHub Copilot CLI installed (https://github.com/github/copilot-cli). Example:
  ```bash
  npm install -g @github/copilot
  copilot auth login
  ```
- Nostr private key (hex) for the runner and at least one allowed operator pubkey.
- Go 1.22+ (for building) or a release binary.

## Configure
Start from the ready-made sample `configs/nostr-copilot.yaml`:
```bash
cp configs/nostr-copilot.yaml config.yaml
```
Fill the placeholders:
- `transports[0].private_key` – runner's nsec hex (never commit this).
- `transports[0].allowed_pubkeys` – operator npub hex list.
- Optional: change relays, working directory, timeouts, and max bytes.

Key knobs:
- `agent.type: copilotcli` – switches to Copilot agent.
- `agent.config.binary` – path to `copilot` if not on PATH (legacy alias: `agent.codex.binary`).
- `agent.config.extra_args` – e.g., `["--allow-all-tools"]` to let Copilot apply edits/execute without interactive approval (legacy alias: `agent.codex.extra_args`).
- `actions` – enable/disable shell and fs actions or tighten allowlists.
- `runner.initial_prompt` – a guardrail shown to Copilot on new sessions.

## Run
```bash
make run                          # uses ./config.yaml
# or point at a custom file
./bin/nostr-codex-runner -config /path/to/config.yaml
```

## Use it
From an allowed Nostr pubkey, DM the runner:
```
/new
List the last 3 git commits in this repo.
/raw ls -la
```
Replies come back over Nostr with the same session id.

## Hardening tips
- Prefer a private relay for production.
- Narrow `actions.shell.allowed` and `actions.*.roots` as much as possible.
- Set `runner.max_reply_chars` to keep DM payloads small.
- Run the health endpoint with `-health-listen 127.0.0.1:8081` and watch logs.

## Troubleshooting
- Copilot timeout: increase `agent.config.timeout_seconds`.
- `copilot: command not found`: point `agent.config.binary` to the absolute path of `copilot`.
- Empty replies: ensure Copilot CLI is enabled for your GitHub account.
