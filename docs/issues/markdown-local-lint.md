# Issue: Ensure markdown lint runs locally

Context

- CI markdownlint now runs across all Markdown (`make lint-md`). Developers should be able to run the same command locally.

Tasks

- Keep `make lint-md` documented as the single-source command for markdownlint-cli2.
- Surface the check in CONTRIBUTING and the PR checklist.
- Optionally add a pre-commit hook snippet in docs/style-guide.md.

Acceptance

- `make lint-md` matches CI coverage/config and is documented.
- PR template reminds contributors to run it.
- CI remains green with the same config.
