# Repo Rename Plan – Issue 3oa.38

Current: `joelklabo/buddy` (rename completed)
Former: `joelklabo/nostr-codex-runner`

Steps

1) **Coordinate timing**
   - Freeze merges during rename window; announce in README and issue tracker.
   - Ensure GitHub Actions secrets carry over to new repo name.

2) **GitHub rename** ✅ completed
   - Verify redirects and Actions after rename.

3) **Module path implications**
   - Plan to update Go module path (covered in issue 3oa.39) immediately after rename to avoid broken `go get`.

4) **Links/badges**
   - Update README badges (CI, release, coverage, staticcheck) to new repo path.
   - Update social preview asset references if path-based.

5) **Releases/download URLs**
   - Update release upload targets in `goreleaser` config and install scripts to new repo URL.

6) **CI/CD**
   - Check workflows for hardcoded `nostr-codex-runner` strings (paths, cache keys, artifact names). Fix after rename.
   - Validate protected branches and required checks settings are preserved.

7) **bd configuration**
   - If bd stores repo prefix, update epic/issue references or note the rename in README/CONTRIBUTING. (bd IDs can stay; repo path change should be documented.)

8) **Docs & external links**
   - Update docs/references to GitHub URLs, release links, install one-liners.
   - Add notice in README/FAQ about old name redirect.

9) **Post-rename verification**
   - Fresh clone using new URL; run `make test`.
   - Verify `go get github.com/joelklabo/buddy/cmd/buddy` works after module update (issue 3oa.39).

Risks/mitigations

- Broken import paths → resolve via module path update and replace directives temporarily.

- Installer scripts fetching old URLs → update scripts and provide redirect notice.

- Cached badges failing → refresh shields.io URLs.
