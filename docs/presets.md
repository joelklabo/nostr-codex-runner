# Buddy Built-in Presets (proposed)

Preset names are positional arguments to `buddy run <preset>`. They can be overridden by dropping a YAML with the same name under `~/.config/buddy/presets/`.

| Name | Purpose | Transport | Agent | Actions | Secrets required | Notes |
|------|---------|-----------|-------|---------|------------------|-------|
| `claude-dm` | Chat via Nostr DMs to Claude/OpenAI-style HTTP backend | nostr | http (Claude/OpenAI) | none by default | API key, relays, allowed pubkeys | Good default demo; wizard should offer it first. |
| `copilot-shell` | Operate host via DMs with Copilot generating commands; executes shell | nostr | copilotcli | shell | Copilot auth, relays, allowed pubkeys | High risk: keep to trusted operators; shell output truncated. |
| `local-llm` | Offline/local LLM answering DMs | nostr (or mock) | http/local endpoint | none | local model endpoint URL or binary path; relays if nostr | Works air-gapped; recommend mock transport for pure local tests. |
| `mock-echo` | No-network demo; echoes prompts | mock | echo | none | none | Good smoke test; use when relays or keys unavailable. |

## Fields
- Each preset YAML declares: `transport` stanza, `agent` stanza, optional `actions`, `logging`, `storage`, `runner` allowlist and prompts.
- Preset metadata (display in `buddy presets`): description, required secrets, default relays, safety notes.

## Search/override order
1) Embedded presets (shipped with binary).
2) `~/.config/buddy/presets/<name>.yaml`
3) `./presets/<name>.yaml` (optional local project overrides).

## Wizard integration
- Wizard should list these presets and pre-fill required fields; selecting a preset sets sensible defaults and then asks only for secrets.
- If the user picks "blank config", wizard skips presets and writes a minimal config.

## Collision handling
- If a user file overrides a built-in preset, `buddy presets <name>` should indicate that it is using the override path.

## Dependency hints
- Presets can declare prerequisites (binaries/env/urls/ports). Run `buddy check <preset>` to see if your environment is ready; add `--json` for scripts.
- Examples: `copilot-shell` expects the `copilot` binary; `claude-dm` will optionally probe `https://api.anthropic.com`; `local-llm` can warn if `127.0.0.1:11434` (ollama default) is closed.

## Open items
- Finalize HTTP agent schema for `claude-dm` (provider field vs URL + headers).
- Decide whether `mock-echo` ships by default or only as a wizard option.
- Confirm action defaults (e.g., disable shell in `claude-dm` to stay safe by default).
