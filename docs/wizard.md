# buddy wizard

Quickest way to produce a working config without editing YAML. Preset-first, with a summary and dependency preflight to catch missing binaries early.

## Usage

```bash
buddy wizard [config-path]
```

- If no path is given, writes `~/.config/buddy/config.yaml` (creates parent dirs).
- If the file exists, prompts before overwrite.

## Flow

1) **Pick a preset** (default `mock-echo`). Shows a one-line summary plus transports/agent/actions in use.
2) **Fill secrets** only for that preset:
   - For nostr presets: relays (defaults), private key (masked), allowed pubkeys.
   - For others, prompts are scoped to what’s missing. `mock-echo` requires no secrets and works fully offline.
3) **Actions**: if shell isn’t already enabled by the preset, ask whether to enable it (warned as high risk).
4) **Dependency preflight**: runs the same checks as `buddy check` (binary/env/file/url/port). Warns and asks before continuing if required deps are missing.
5) **Dry-run?** preview without writing.

## What it writes

- Transport: nostr with your relays and keys (or mock if chosen)
- Agent: chosen type
- Actions: readfile always; shell if enabled
- Projects: default project at `.`
- Storage: `~/.buddy/state.db`

## After running

- Output shows the config path.
- Start the runner:
  
  ```bash
  buddy run -config ~/.config/buddy/config.yaml
  ```

## Notes

- Uses masked prompts; secrets are not echoed.
- Writes to buddy paths by default; pass a custom path to override.
- Same config search order as `buddy run` applies when you start the runner.
- To bypass prereq checks when experimenting, re-run wizard and decline to continue if missing deps (or add them and rerun).
- Privacy & safety: the wizard only prints summaries (no secrets), asks before overwriting files, and stores secrets locally in your config. It does not phone home or emit telemetry.
