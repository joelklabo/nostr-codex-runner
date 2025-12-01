# Issue: Markdown Cleanup Backlog

Context

- markdownlint now runs across all Markdown via `make lint-md` (markdownlint-cli2@0.19.1).
- Legacy violations have been cleared; keep the repo clean and document any future exclusions explicitly.

Tasks

- Keep new/edited docs lint-clean or add scoped rule disables with rationale.
- If an exclusion is needed, add it to `.markdownlint.yaml` with a comment.

Acceptance

- markdownlint covers the repo with zero violations.
- Any exclusions are documented in `.markdownlint.yaml` with justification.
