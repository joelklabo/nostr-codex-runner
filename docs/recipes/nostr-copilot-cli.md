# Recipe: Nostr transport + GitHub Copilot CLI agent

This walks through running the pluggable runner with:
- **Transport**: Nostr DMs.
- **Agent**: GitHub Copilot CLI (`gh copilot suggest`).
- **Actions**: shell, readfile, and writefile.

## Prerequisites
- GitHub CLI (`gh`) installed and authenticated.
- Copilot CLI enabled: `gh extension install github/gh-copilot && gh copilot setup`.
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
- `agent.codex.binary` – path to `gh` if not on PATH.
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
- Copilot timeout: increase `agent.codex.timeout_seconds`.
- `gh: command not found`: point `agent.codex.binary` to the absolute path of `gh`.
- Empty replies: ensure Copilot CLI is enabled for your GitHub account.
