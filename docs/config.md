# Configuration Reference

Scope: key fields for buddy configs and presets. See `config.example.yaml` for the full annotated example.

## Search order (run)
1) Explicit path argument to `buddy run <config>` (if not a preset name)
2) `./config.yaml`
3) `~/.config/buddy/config.yaml`

## Preset search order
1) `~/.config/buddy/presets/<name>.yaml` (user overrides)
2) `./presets/<name>.yaml` (project overrides)
3) Embedded presets shipped with the binary (fallback)

## Top-level fields
| Field | Type | Default | Notes |
|-------|------|---------|-------|
| `transports` | list | required | One or more transports (nostr, mock; others pluggable). |
| `agent` | object | required | Model backend selection. |
| `actions` | list | [] | Host capabilities (shell, readfile, writefile). |
| `runner` | object | defaults | Allowlist, session timeouts, initial prompt. |
| `storage` | object | `~/.buddy/state.db` | BoltDB path. |
| `logging` | object | level=info, format=text | Supports json, optional file. |

## Runner
- `allowed_pubkeys` (list, required for nostr): who can control the runner.
- `session_timeout_minutes` (int, default 60): idle timeout.
- `initial_prompt` (string, optional): prepended once per new session.
- `max_reply_chars` (int): truncate replies.
 - `profile_name`/`profile_image`: optional display fields.

## Transport: nostr
| Field | Type | Default/Notes |
|-------|------|---------------|
| `type` | string | `nostr` |
| `id` | string | unique transport id |
| `relays` | list | e.g., `wss://relay.damus.io` |
| `private_key` | hex string | required (nsec hex) |
| `allowed_pubkeys` | list | should match runner allowlist |

## Transport: mock
- `type: mock`
- No secrets. Good for smoke tests/offline.

## Agent options
- **http** (Claude/OpenAI style)
  - `type: http`
  - `config.base_url`, `config.model`, `config.api_key` (secret), `config.timeout_seconds`.
- **copilotcli**
  - `type: copilotcli`
  - `config.binary` (default `copilot`), `working_dir`, `timeout_seconds`, `extra_args`.
- **codexcli**
  - `type: codexcli`
  - `config.binary`, `working_dir`, `timeout_seconds`, `extra_args`.
- **echo**
  - `type: echo` (offline/testing).

## Actions
- **shell**: `workdir`, `timeout_seconds`, `max_output`.
- **readfile**: `roots` allowlist.
- **writefile**: `roots` allowlist, `allow_write`, `max_bytes`.

## Storage
- `storage.path`: BoltDB file path (default `~/.buddy/state.db`).

## Logging
- `logging.level`: `debug|info|warn|error`.
- `logging.format`: `text|json`.

## Preset schema additions
- `meta.description`: short description shown in `buddy presets`.
- `meta.secrets`: list of required secrets (for display only).
- `meta.safety`: notes (e.g., shell risk).

## Defaults and validation tips
- If no transports are provided, config defaults to a nostr transport using runner keys.
- If transports exclude nostr, runner keys/allowlist fall back to `"mock"` to keep validation green; provide real keys for nostr.
- If no actions are provided, a `readfile` action is auto-added with roots rooted at the config directory.
- Ensure nostr keys are hex, 64 chars (npub values are normalized to hex).
- Keep allowlists non-empty when transports enable actions.
- Avoid enabling `shell` in public/unknown environments.
