# Release Checklist (buddy)

Run before tagging:

- [ ] `make lint-md`
- [ ] `go test ./...`
- [ ] `make brew-check` (macOS; validates tap install/test)
- [ ] `make man` (ensure docs/man/buddy.1 is fresh)
- [ ] `git status` clean; changelog updated

Tag/release:

- [ ] Create tag `vX.Y.Z`
- [ ] `git push origin --tags`
- [ ] Wait for GitHub Actions: CI, gosec, govulncheck, coverage, markdownlint, homebrew-check, docker builds
- [ ] Verify release artifacts (tar.gz, checksums, man page) on GitHub Releases

Post-release:

- [ ] Update Homebrew tap if required (CI should handle)
- [ ] Smoke: `buddy run mock-echo` from release binary
- [ ] Update README badges if versions change
