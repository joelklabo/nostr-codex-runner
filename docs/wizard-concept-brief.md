# buddy wizard – Concept Brief (Issue 3oa.17)

## Goal

Get a first usable DM session in under 3 minutes with zero YAML editing. Defaults are safe (mock or read-only), secrets stay local, and users see clear next steps.

## Scope

- Transport: start with Nostr or mock. Nostr prompts only when selected.

- Agents: Claude/OpenAI HTTP, Copilot CLI, echo/local. Present as presets to avoid matrix explosion.

- Actions: readfile on by default; shell is opt-in with a strong warning; writefile optional.

- Outputs: write a single `config.yaml`, print next commands (`buddy run <preset>`, `buddy check <preset>`), and show dep preflight results.

- UX: CLI (survey prompts), no TUI/colors required for MVP.

## Success metrics

- Time-to-first-DM: <3 minutes for a new user using mock-echo or claude-dm.

- Error rate: dep preflight catches missing binaries/ports for >=90% of first runs.

- Abandonment: overwrite prompts and shell warning reduce accidental risky runs; secrets never echoed.

## Flow (implemented)

1) Choose preset (default mock-echo), show summary (transports/agent/actions).
2) Ask only relevant secrets for that preset.
3) Shell opt-in if not already enabled by preset.
4) Dependency preflight; allow continue on WARN, stop or confirm on MISSING.
5) Write config; print next commands.

## Guardrails

- Mask secrets, no telemetry, no secret logging.

- Confirm before overwriting existing file.

- Shell always framed as high risk.

## Extensibility

- Registry-driven options (`SetRegistry`) for transports/agents/actions/presets; plugins can inject options and tests can override.

## Open questions

- Do we add a “run now?” prompt after writing? (currently just prints commands)

- Add minimal analytics locally (count prompt steps) without phoning home? likely no.

- Should we keep mock-echo as the default preset long term or rotate to claude-dm once creds are present?
