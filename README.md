# buddy — friendly transport → agent → actions runner

Binary-first, wizard-assisted CLI that turns Nostr DMs (and other transports) into flexible agent pipelines. Pick a preset or run the wizard and get a working config in minutes.

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

## Quick start — 3 steps

1) **Install**: `brew install joelklabo/tap/buddy` (or the curl script above).
2) **Pick a preset or wizard**:
   - `buddy run mock-echo` (zero secrets, offline)
   - `buddy run claude-dm` (Nostr → Claude/OpenAI HTTP)
   - `buddy run copilot-shell` (Copilot + shell; trusted operators only)
   - or `buddy wizard` to generate `~/.config/buddy/config.yaml` interactively.
3) **DM it**: from an allowed pubkey, send `/new` and try a prompt. Stop with Ctrl+C.

## What it does
- **Transports**: Nostr DMs (default) or mock; pluggable others.
- **Agents**: Claude/OpenAI-style HTTP, Copilot CLI, local codexcli/echo.
- **Actions**: shell, readfile, writefile; add your own.
- Keeps session context, enforces allowlists, exposes optional health/metrics.

## CLI surface
```
buddy run <preset|config>      start the runner
buddy wizard [config-path]     guided setup; supports --dry-run
buddy presets [name]           list built-ins or show details
buddy check <preset|config>    verify dependencies (use --json for machine output)
buddy version                  show version
buddy help                     show help
```

## Config search order
1) Positional/flag: `buddy run <preset|path>` (flag `-config` still works)
2) `./config.yaml`
3) `~/.config/buddy/config.yaml`

## Docs
- Docs index: `docs/index-ia.md` (user vs contributor landing) – to be published
- Quick install: `docs/quick-install-ia.md`
- Use cases: `docs/use-cases.md`
- Config reference: `docs/config.md`
- Wizard: `docs/wizard.md`
- Presets: `docs/presets.md`

## Want to help?
- Read `CONTRIBUTING.md` (setup, style, test expectations) and peek at the `buddy-3oa` epic in `bd list`.
- Pick a first issue: docs polish, new preset, or improving `buddy help`. If unsure, open a discussion and we’ll pair you up.
- One commit per issue, tests before push. Friendly reviews welcome.

## Contributing
- Issues tracked with `bd` (epic `buddy-3oa`); one commit per issue.
- Go 1.24+, `go test ./...` before pushing.
- See `CONTRIBUTING.md` and `docs/style-guide.md`.
- New here? Start with `buddy run mock-echo`, then pick a tiny doc fix or preset tweak and open a PR. Friendly reviews welcome.

## FAQ
- **Do I need Go to use it?** No. Install a release (brew or script) and run presets; build from source is optional.
- **Is shell safe?** Shell is powerful and risky—use `mock-echo` or `claude-dm` first, and keep `allowed_pubkeys` tight before enabling shell.
- **Where do configs live?** By default `~/.config/buddy/config.yaml`; presets are built-in but can be overridden under `~/.config/buddy/presets/`.
- **Is there a short alias?** We ship only `buddy` to avoid colliding with other tools; if you want `bud`, create your own shell alias or symlink.

## License
MIT
