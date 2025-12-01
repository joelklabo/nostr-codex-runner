# Packaging & Releases Plan – Issue 3oa.28

Goal: ship buddy binaries via GitHub Releases with checksums/signing, Homebrew tap, and optional install script.

Deliverables
- goreleaser config updated for `buddy` binary (+ `nostr-buddy` symlink).
- Release artifacts: tar.gz per OS/arch (darwin/linux, amd64/arm64), SHA256 checksums, optional cosign signatures.
- Homebrew tap formula updated/created.
- CI release workflow wired to tags.

Steps
1) Update/author `.goreleaser.yml`
   - Build matrix: darwin/linux, amd64/arm64.
   - Binary name `buddy`; post-build symlink `nostr-buddy` in archives.
   - Archive naming: `buddy_<version>_<os>_<arch>.tar.gz`.
   - Checksums file; optional `cosign sign-blob` if keys available.
   - Include license/README in archives? keep small; include `config.example.yaml` and presets? (TBD)

2) Release workflow
   - Adjust `.github/workflows/release.yml` to trigger on tags `v*` in new repo.
   - Use Go 1.24.x; cache modules; run `goreleaser release --clean`.
   - Publish artifacts and checksums.

3) Homebrew tap
   - Update existing tap or create new `homebrew-buddy` tap.
   - Template formula to fetch tarball and install `buddy`; add `nostr-buddy` symlink.
   - Test `brew install joelklabo/tap/buddy` on macOS arm64/amd64.

4) Install script
   - Ensure `scripts/install.sh` fetches latest release from new repo, verifies checksum, installs to `/usr/local/bin` or `~/.local/bin`, and can choose alias when collision detected.

5) Verification checklist per release
   - `buddy --version` matches tag.
   - Checksums match downloaded artifact.
   - `buddy run mock-echo` works on macOS and Linux.
   - Homebrew install/uninstall works.

6) Documentation updates
   - README quick install uses brew and script URLs post-rename.
   - Quick-install doc references checksum and alias handling.

Open questions
- Whether to publish detached cosign signatures; if yes, define key management.
- Whether to embed presets/config example in archives vs separate download.
- Decide on release cadence (align with changelog policy issue 3oa.11).
