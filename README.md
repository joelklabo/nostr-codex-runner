# buddy – pluggable transport → agent → actions CLI

> Binary-first, wizard-assisted runner for Nostr DMs (and more). Pick a preset or run the wizard, get a working config in minutes.

[![CI](https://github.com/joelklabo/buddy/actions/workflows/ci.yml/badge.svg)](https://github.com/joelklabo/buddy/actions/workflows/ci.yml)
[![Release](https://github.com/joelklabo/buddy/actions/workflows/release.yml/badge.svg)](https://github.com/joelklabo/buddy/actions/workflows/release.yml)
[![Coverage](https://github.com/joelklabo/buddy/actions/workflows/coverage.yml/badge.svg)](https://github.com/joelklabo/buddy/actions/workflows/coverage.yml)
[![Staticcheck](https://github.com/joelklabo/buddy/actions/workflows/staticcheck.yml/badge.svg)](https://github.com/joelklabo/buddy/actions/workflows/staticcheck.yml)
[![Go](https://img.shields.io/badge/go-1.24%2B-00ADD8)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](#license)

## Quick install

```bash
brew install joelklabo/tap/buddy           # Homebrew (macOS/Linux)
# or
curl -fsSL https://get.buddy.sh | sh       # script installer (downloads release, verifies checksum)
```

## Quick start (choose one)

1) **Wizard (guided setup)**
```bash
buddy wizard                               # writes ~/.config/buddy/config.yaml (masked secrets)
buddy run -config ~/.config/buddy/config.yaml
```

2) **Preset (no edits)**
```bash
buddy run mock-echo                        # offline smoke test
buddy run claude-dm                        # Nostr DM to Claude/OpenAI style HTTP
buddy run nostr-copilot-shell              # Copilot + shell action (trusted operators only)
```

3) **Explicit config path**
```bash
buddy run path/to/config.yaml              # argv beats env/cwd
```

Expected output: a short startup banner then streaming logs; stop with Ctrl+C.

## What it does
- **Transport**: nostr DMs (default) or mock; more transports are pluggable.
- **Agent**: codexcli, copilotcli, HTTP (Claude/OpenAI style), echo, local endpoints.
- **Actions**: shell, readfile, writefile (extendable).
- Keeps session context, enforces allowlists, exposes optional health/metrics.

## CLI surface
```
buddy run <preset|config>      start the runner
buddy wizard [config-path]     guided setup; supports --dry-run
buddy presets [name]           list built-ins or show details
buddy version                  show version
buddy help                     show help
```

## Config search order
1) argv `-config` path (or positional in future)
2) `./config.yaml`
3) `~/.config/buddy/config.yaml`
4) Legacy: `NCR_CONFIG` env and `~/.config/nostr-codex-runner/config.yaml` (supported with warning for one release)

## Docs
- Docs index: `docs/index-ia.md` (user vs contributor landing) – to be published
- Quick install: `docs/quick-install-ia.md`
- Use cases: `docs/use-cases.md`
- Config reference: `docs/config.md`
- Wizard: `docs/wizard.md`
- Presets: `docs/presets.md`

## Contributing
- Issues tracked with `bd` (epic `nostr-codex-runner-3oa`); one commit per issue.
- Go 1.24+, `go test ./...` before pushing.
- See `CONTRIBUTING.md` and `docs/style-guide.md`.
- New here? Try `buddy run mock-echo`, then open a PR with your first improvement. Friendly reviews welcome.

## License
MIT
