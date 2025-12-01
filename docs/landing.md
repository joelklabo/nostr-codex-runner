# Buddy Landing (lean content draft)

Purpose: Single-page content (can be GitHub Pages or README front) highlighting install, presets, wizard, and contributing.

## Hero
- Headline: “Run AI agents over Nostr in minutes.”
- Subhead: “Binary-first CLI with presets, wizard, and safe defaults.”
- Buttons: `brew install joelklabo/tap/buddy`, `buddy wizard`, `buddy run mock-echo`.

## Proof/Badges
- CI, Release, Coverage badges (reuse README).

## Quickstart (3 steps)
1) Install (brew or script).
2) Pick a preset: mock-echo, claude-dm, copilot-shell, local-llm.
3) DM `/new` from an allowed pubkey (or run offline with mock).

## Why Buddy
- Pluggable transports → agents → actions.
- Presets included; wizard writes config.
- Dependency checks built-in (`buddy check`).
- Friendly logs, health/metrics endpoints.

## Presets
- Cards for each preset with description and `buddy run <name>` command.

## Safety
- Shell is opt-in and warned.
- Allowed pubkeys and dep checks.
- No telemetry; secrets local.

## Contribute
- Link to CONTRIBUTING.md, bd workflow, and first-issue pointers.

## Footer
- Links: README, Docs index, FAQ, Security, Changelog, Releases.
