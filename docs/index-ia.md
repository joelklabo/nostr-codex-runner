# Buddy Docs Landing (Information Architecture)

Goal: one landing page that routes **Users** and **Contributors** quickly, keeping README lean. This IA will become `docs/index.md`.

## Top-level structure
- Hero: one-liner + primary CTAs: `Quickstart (Users)` and `Contribute (Developers)`.
- Secondary nav: `Presets`, `Wizard`, `Config`, `Recipes`, `FAQ`, `Security`, `Changelog`.

## User track (getting running fast)
1) Quickstart
   - 3-step install and run (`brew install buddy` / `curl | sh`, `buddy presets`, `buddy wizard` or `buddy run <preset>`).
   - Expected first DM/response; config path emitted.
2) Presets
   - Table of built-ins (`copilot-shell`, `claude-dm`, `local-llm`) with one-line purpose and run command.
3) Wizard
   - Why/when to use; sample session transcript; config output location.
4) FAQ / Troubleshooting
   - Install/path issues, relay/auth, BoltDB lock, missing keys, collision with other `buddy` CLIs.
5) Recipes
   - Links to task-specific guides (Nostr + Claude, Local LLM offline, Local LLM mock override, Custom action trigger, Systemd/service, Docker usage).
6) Security quick note
   - Allowlist, secret logging stance; link to full Security doc.
7) Releases / Changelog
   - Link to latest release and changelog policy.

## Contributor track (building/adding)
1) Contributing guide
   - bd workflow, one commit per issue, Go 1.22+, testing commands.
2) Local dev setup
   - go mod download, lint/test targets, sample data paths, BoltDB locations.
3) Architecture overview
   - Diagram link; brief text on transport→agent→actions, storage, metrics.
4) Extending buddy
   - How to add transports/agents/actions; how to add CLI subcommands and presets; registry/code pointers.
5) Config reference
   - Full schema, defaults, search order; preset schema.
6) Wizard internals
   - Registry/extensibility notes; tests/goldens.
7) Release process
   - Packaging/release pipeline (goreleaser), Homebrew tap update, QA matrix.
8) Style guide
   - Docs style + code style links; templates.

## Footer links
- README (short), Releases, Security policy, Code of Conduct, License.

## Open items / placeholders
- Publish docs under the renamed `buddy` repo and confirm installer URL (`https://get.buddy.sh`) is stable.
- Finalize preset list and commands when CLI spec stabilizes (current set: claude-dm, copilot-shell, local-llm, mock-echo).
