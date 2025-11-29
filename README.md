# Nostr Codex Runner

Always-on service that listens for Nostr DMs from trusted pubkeys and pipes their content into `codex exec`, preserving Codex session IDs so you can continue conversations via Nostr.

## Features
- Subscribes to encrypted DMs (`kind 4`) addressed to the runner key and authored by an allowlist of pubkeys.
- Command mini-DSL:
  - `/new [prompt]` reset active session; optional prompt starts a fresh Codex session.
  - `/use <session-id>` switch to an existing Codex session.
  - `/status` show the active session.
  - `/help` usage recap.
  - Anything else is treated as a prompt and run in the active session (or creates a new one).
- Persists active session per sender (Bolt DB) and expires it after a configurable idle window.
- Replies via Nostr DM with the Codex session ID and last agent message (truncated to `max_reply_chars`).

## Quick start
1) Copy `config.example.yaml` to `config.yaml` and fill in secrets:
   - `runner.private_key`: hex nostr sk (nsec) for the runner.
   - `runner.allowed_pubkeys`: pubkeys that are allowed to issue commands.
   - Adjust relays, working directory for Codex, etc.
2) Run locally:
```bash
make run             # uses config.yaml
# or
CONFIG=path/to/config.yaml ./scripts/run.sh
```
3) Send an encrypted DM from an allowed pubkey to the runner's pubkey. Example payloads:
```
/new
/new Write a Go HTTP server that echoes requests.
List the last 5 git commits in this repo.
/status
```

## Background service options (macOS-friendly)
- **tmux/screen:** `tmux new -s codex-runner 'cd /Users/honk/code/nostr-codex-runner && make run'`
- **launchd:** create `~/Library/LaunchAgents/com.honk.nostr-codex-runner.plist`:
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
Load it with `launchctl load ~/Library/LaunchAgents/com.honk.nostr-codex-runner.plist`.

## Implementation notes
- Go 1.25 with [`go-nostr`](https://github.com/nbd-wtf/go-nostr) for relay connectivity and NIP-04 encryption.
- Lightweight persistence via BoltDB (`storage.path`, default `state.db`).
- Codex integration uses `codex exec --json` and captures `thread_id` as the Codex session id. Replies stream the last `agent_message` text.
- Session inactivity timeout (`session_timeout_minutes`) automatically discards stale sessions.

## Security tips
- Keep `config.yaml` out of version control (already in `.gitignore`).
- Use relays you trust; consider running a private relay.
- Restrict `allowed_pubkeys` to yourself while testing.

## bd workflow
- Repo already has an epic (`nostr-codex-runner-2zo`) and feature issue (`nostr-codex-runner-2zo.1`).
- Follow-up work should be recorded via `bd create --parent nostr-codex-runner-2zo ...`, and commits should reference/close the relevant issue.
