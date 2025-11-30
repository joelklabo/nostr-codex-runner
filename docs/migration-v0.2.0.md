# Migration to plugin architecture (v0.2.0)

Breaking-ish changes:
- Config now supports `transports`, `agent`, and `actions` arrays/objects. Old top-level fields (`relays`, `codex`, `runner.*`) still work via defaults/shims.
- Default transport is built from legacy keys if `transports` is omitted.
- Default agent is `codexcli`; default actions include `shell` if none are configured.

What to check:
1) If you used custom Nostr keys: ensure `transports[0].private_key` and `allowed_pubkeys` are set (or keep legacy fields).
2) If you added shell allow/deny lists: move them under the `shell` action config.
3) Storage path unchanged (`storage.path`).

Examples:
- See `configs/nostr-codex-shell.yaml` for the current Nostr+Codex setup.
- See `configs/mock-echo.yaml` for local testing with mock transport + echo agent.

Runtime expectations:
- Actions now honor allowlists and per-action timeouts; action calls are audited in Bolt.
- Session timeout and initial prompt are preserved from legacy runner settings.

If something fails:
- Delete `state.db` only if you must; schema is backward-compatible (new buckets added for history/audit).
