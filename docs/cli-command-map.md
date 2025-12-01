# buddy CLI Command Map (argument-first)

Design target: minimal flags, positional arguments, fast muscle memory, clear help. All commands assume the binary name is `buddy`; aliases like `nostr-buddy` or `bud` may point to the same binary.

## Commands
- `buddy run <preset|config>`
  - If the argument matches a shipped preset name, load that preset and merge user overrides.
  - Otherwise treat as a file path to a config YAML (relative or absolute).
  - Flags: `-config <path>` (default search order), `-skip-check`, `-health-listen`, `-metrics-listen`.
  - Output: startup summary (transport, agent, actions, config path), then structured logs.
  - Exit codes: 0 clean exit, 2 config/preset not found, 3 validation error, 4 runtime fatal.

- `buddy wizard [config-path]`
  - Interactive guided setup; optional positional path to write config (default: `~/.config/buddy/config.yaml`).
  - Prompts: transport/relays, keys (hidden input), allowed pubkeys, agent choice, actions, preset selection, optional smoke test.
  - Flags avoided; offer a `--dry-run` only if necessary for safety (TBD).

- `buddy presets`
  - Lists built-in presets with one-line description and required secrets.
  - `buddy presets <name>` shows details for a single preset (inputs, actions, sample command).

- `buddy check <preset|config>`
  - Verifies declared dependencies (binary/env/file/url/port/relay/dirwrite).
  - Flags: `-config <path>` (same search order), `-json` for machine-readable output.

- `buddy init-config [path]`
  - Writes the bundled example config to `./config.yaml` (or provided path) if missing.

- `buddy help`
  - Short usage and pointers; `buddy help run|wizard|presets` for detail.

- `buddy version`
  - Prints semantic version + git commit; optional `--json` (only flag currently allowed) for scripting.

## Global behaviors
- Config search order (for run):
  1) Explicit positional path (if not a preset name)
  2) `./config.yaml`
  3) `~/.config/buddy/config.yaml`
- Preset search order:
  1) Built-in embedded presets
  2) `~/.config/buddy/presets/<name>.yaml`
  3) `./presets/<name>.yaml` (optional)
- Aliases: provide `nostr-buddy` symlink; `bud` alias opt-in because of collisions.
- Logging: default human-readable; `--json` only for `version` and maybe `run --json-logs`? (TBD; keep defaults simple.)
- Environment variables: opt-in for advanced users (e.g., `BUDDY_CONFIG`, `BUDDY_PRESET_PATH`), but not required for basic use.

## Help copy (draft)
```
buddy run <preset|config>      start the runner from a preset or YAML config
buddy check <preset|config>    verify dependencies (json optional)
buddy wizard [config-path]     guided setup; writes a config
buddy presets [name]           list built-in presets or show one
buddy init-config [path]       write example config if missing
buddy help [cmd]               show help
buddy version                  show version info
```

## Open questions
- Do we expose a `--json-logs` flag on `run` or keep logs text-only by default?
- Should `wizard` support `--no-save` for privacy testing?
- Confirm alias install strategy and how Homebrew formula handles it.
