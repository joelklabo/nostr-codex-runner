# Issue: Enforce Markdown Quality and Prevent Regressions

Context

- A formatting lapse (see codex-clipboard-4P6sle.png) suggests we lacked or bypassed markdown linting/preview.
- We added a basic `markdownlint` workflow (`.github/workflows/markdownlint.yml`) and config (`.markdownlint.yaml`) allowing our inline HTML.

Tasks

- Run markdownlint across the repo to ensure zero violations.
- Add the workflow badge to README or CI docs so failures are visible.
- Wire lint into the main CI pipeline (required check on PRs).
- Document the lint command for contributors in `CONTRIBUTING.md`.

Acceptance

- Lint job runs on push/PR, is required for merge, and repo is clean.
- README/CONTRIBUTING mention how to run lint locally.
- Future PRs failing markdown lint are blocked until fixed.
