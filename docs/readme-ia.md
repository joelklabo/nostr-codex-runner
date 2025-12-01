# README information architecture (buddy CLI)

Goal: sub-400 word README that sells the binary-first buddy CLI, shows how to install and try it in 2 minutes, and defers details to docs.

## Proposed sections (order)
1) **Hero/what** – one sentence: "buddy is a pluggable transport→agent→actions runner (Nostr-first) with presets and a guided wizard."
2) **Quick install** – two options:
   - Homebrew: `brew install buddy` (or tap path once defined).
   - Script: `curl -fsSL https://get.buddy.sh | sh` (checksum note, installs to ~/.local/bin or /usr/local/bin).
   - Note OS/arch supported; alias collision warning (Buddy.Works).
3) **Quick start** – 3 commands, argument-only:
   - `buddy presets` (list built-ins).
   - `buddy wizard` (generate config interactively) or `buddy run nostr-copilot-shell` (preset shortcut).
   - Show expected first DM/response flow and config path emitted.
4) **Use cases** (links to docs/examples):
   - Nostr DM → Claude/OpenAI (preset `claude-dm`).
   - Local model / offline flow (preset `local-llm`).
   - Custom action trigger / shell (`nostr-copilot-shell`).
5) **Features** (bullet): presets, wizard, pluggable transports/agents/actions, session tracking, replay protection, logging/metrics.
6) **CLI surface** (mini table):
   - `buddy run <preset|config.yaml>` – start runner.
   - `buddy wizard` – guided setup; writes config.
   - `buddy presets` – list/describe built-ins.
   - `buddy help` – short help; `--version`.
   - Aliases: optional `bud`, `nostr-buddy` (mention collision warning).
7) **Configuration** – 1 paragraph pointing to `docs/config.md` and preset schema; search order (argv, cwd, ~/.config/buddy/config.yaml).
8) **Contributing** – link to CONTRIBUTING (CLI-first, bd issues, one commit per issue).
9) **License/links** – MIT, releases, docs index, changelog.

## Tone & layout notes
- Keep first screen above the fold: hero + quick install + quick start.
- Avoid long plugin lists; link to plugin catalog instead.
- Prefer short sentences and command blocks; trim adjectives.
- Mention security briefly (allowlist, no secret logging) with link to Security doc.

## Open questions
- Final install URL/tap name after repo rename.
- Whether to default to `buddy` binary or ship `nostr-buddy` alias prominently in README.
