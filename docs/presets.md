# Buddy Built-in Presets

Preset names are positional arguments to `buddy run <preset>`. They can be overridden by dropping a YAML with the same name under `~/.config/buddy/presets/`.

|Name|Purpose|Transport|Agent|Actions|Secrets required|Notes|
|---|---|---|---|---|---|---|
|`claude-dm`|Chat via Nostr DMs to Claude/OpenAI-style HTTP backend|nostr|http (Claude/OpenAI)|none|API key, relays, allowed pubkeys|Good default demo; wizard should offer it first.|
|`copilot-shell`|Operate host via DMs with Copilot generating commands; executes shell|nostr|copilotcli|shell|Copilot auth, relays, allowed pubkeys|High risk: keep to trusted operators; shell output truncated.|
|`local-llm`|Offline/local LLM answering DMs|nostr|http/local endpoint|none|local model endpoint URL or binary path; relays if nostr|Works air-gapped; override transport to `mock` if you maintain a local build.|
|`mock-echo`|No-network demo; echoes prompts|mock|echo|none|none|Good smoke test for automated tests; mock transport does not accept live DMs.|

## Fields

- Each preset YAML declares: `transport` stanza, `agent` stanza, optional `actions`, `logging`, `storage`, `runner` allowlist and prompts.

- Preset metadata (display in `buddy presets`): description, required secrets, default relays, safety notes.

## Search/override order

1) `~/.config/buddy/presets/<name>.yaml` (user override, if present).
2) `./presets/<name>.yaml` (project override, if present).
3) Embedded presets (shipped with the binary).

## Wizard integration

- Wizard should list these presets and pre-fill required fields; selecting a preset sets sensible defaults and then asks only for secrets.

- If the user picks "blank config", wizard skips presets and writes a minimal config.

## Collision handling

- If a user file overrides a built-in preset, `buddy presets <name>` should indicate that it is using the override path.

## Dependency hints

- Presets can declare prerequisites (binaries/env/urls/ports). Run `buddy check <preset>` to see if your environment is ready; add `--json` for scripts.

- Examples: `copilot-shell` expects the `copilot` binary; `claude-dm` will optionally probe `https://api.anthropic.com`; `local-llm` can warn if `127.0.0.1:11434` (ollama default) is closed.

## Notes

- Shipped presets leave secrets empty and `allowed_pubkeys` blank; run `buddy wizard` or provide an override to fill them before `buddy run <preset>`.

- `buddy presets <name> --yaml` shows the effective YAML (built-in or override).

- Dependency hints are surfaced via `buddy check <preset>`; use `--json` for scripting.

- For offline local-llm with mock transport, see `docs/recipes/local-llm-mock.md`.
