# Wizard Concept Brief (buddy) – Issue 3oa.17

Goals

- Reduce "first successful DM" time to under 5 minutes with no prior repo knowledge.

- Generate a valid config (and optionally select a preset) without editing YAML manually.

- Explain choices briefly so new users don’t need to read full docs.

Scope

- CLI (no TUI) interactive wizard invoked via `buddy wizard [config-path]`.

- Outputs a config file and prints next command (e.g., `buddy run claude-dm`).

- Optional smoke test prompt at the end (start runner with mock transport or chosen preset).

- Secrets (keys, tokens) entered via masked prompts; never logged.

Success metrics

- Time-to-first-response median <5 minutes for new users.

- Wizard-generated configs pass validation and start without edits in >90% of runs.

- Less than 1 secret exposure per 100 runs (no echo/backscroll leaks).

High-level flow

1) Welcome + explain what will be written and where.
2) Transport selection (default: nostr). If nostr: ask for relays (prefill defaults), private key (hidden), allowed pubkeys. Offer mock transport for offline test.
3) Agent selection: options Claude/OpenAI HTTP, Copilot CLI, local LLM (HTTP/local). Collect necessary fields.
4) Actions: default none; offer shell/readfile/writefile; warn about shell risk.
5) Preset alignment: allow user to pick a preset to base on (claude-dm, copilot-shell, local-llm, mock-echo) which auto-fills choices above.
6) Config path confirmation (default `~/.config/buddy/config.yaml`); create directories as needed.
7) Write config; print summary and next-step commands (`buddy run <preset>` or `buddy run <config>`).
8) Optional smoke test: start runner now? If yes, run with chosen preset/ config and stream logs until Ctrl+C.

Out of scope for MVP

- TUI/colored animations

- Telemetry/analytics collection

- Non-interactive flags beyond optional `--dry-run` or `--no-save`

Risks/mitigations

- Secret leakage in logs → keep prompts masked, avoid echo; ensure error paths don’t print secrets.

- Misconfig due to defaults → pick safe defaults (no shell action unless confirmed, allowlist required, nostr relays prefilled but editable).

- Path collisions → warn if config file exists; prompt to overwrite or choose a new path.
