# README Details

## CLI surface

```text
buddy run <preset|config>      start the runner
buddy wizard [config-path]     guided setup; supports --dry-run
buddy presets [name]           list built-ins or show details
buddy check <preset|config>    verify dependencies (use --json for machine output)
buddy version                  show version
buddy help                     show help
buddy init-config [path]       copy example config if missing
```

## Config search order

1) Positional/flag: `buddy run <preset|path>` (flag `-config` still works; env `BUDDY_CONFIG` preferred)
2) `./config.yaml`
3) `~/.config/buddy/config.yaml`

## Docs index

- Docs landing: `docs/index-ia.md`
- Quick install: `docs/quick-install-ia.md`
- Use cases: `docs/use-cases.md`
- Config reference: `docs/config.md`
- Wizard: `docs/wizard.md`
- Presets & overrides: `docs/presets.md`
- Offline local LLM (mock): `docs/recipes/local-llm-mock.md`
- CLI map/help: `docs/cli-command-map.md`
- Release QA: `docs/release-qa.md`

## FAQ (short)

- **Do I need Go to use it?** No. Install a release (brew or script) and run presets; building is optional.
- **Is shell safe?** Shell is powerful—start with `mock-echo` or `claude-dm`; keep `allowed_pubkeys` tight before enabling shell.
- **Where do configs live?** By default `~/.config/buddy/config.yaml`; presets can be overridden under `~/.config/buddy/presets/`.
- **Alias?** We ship `buddy`. Create your own alias if you want `bud`.
- **Windows?** Not yet. macOS/Linux (amd64/arm64). WSL works with the Linux binary.
