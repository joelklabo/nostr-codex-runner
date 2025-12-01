# Issue: Verify Homebrew Tap in CI

Goal
- Prove the README claim “Homebrew tap `joelklabo/tap` is published by this repo’s releases—no extra setup needed” by running an automated check.

Proposal
- Add a macOS CI job (nightly + on release PRs) that:
  1) `brew tap joelklabo/tap`
  2) `brew info buddy` (asserts formula exists)
  3) `brew install joelklabo/tap/buddy` (optionally with `--build-from-source` to avoid cached bottles drift)
  4) `buddy --version` to ensure the binary runs
- Fail the workflow if any step fails; keep the job optional on PRs if it’s slow.

Acceptance
- Workflow added and passing on main; documented in `docs/release-qa.md`.
- README line remains accurate and linked to the CI job badge/log.
