# Issue: Ensure markdown lint runs locally

Context

- CI markdownlint now runs on key docs, but developers may miss formatting issues until push.

Tasks

- Add a helper script or make target to run markdownlint-cli2 with repo config.
- Document local command in CONTRIBUTING (if not already) and surface in PR template checklist.
- Optionally add a pre-commit hook snippet in docs/style-guide.md.

Acceptance

- One-line local lint command available (script/Makefile target) and documented.
- PR template mentions running markdownlint.
- CI remains green with the same config.
