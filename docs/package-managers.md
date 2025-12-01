# Package Manager Updates – Issue 3oa.40

Homebrew tap

- Update or create formula for `buddy` in `joelklabo/homebrew-tap` (or new tap). Points to new repo release artifacts.

- Install stanza: `bin.install "buddy"` and `bin.install_symlink "buddy" => "nostr-buddy"` if alias shipped.

- Test on macOS arm64/amd64: install, version, uninstall.

Checksum URLs

- goreleaser needs to publish SHA256 checksums; formula references the correct filename pattern `buddy_<version>_<os>_<arch>.tar.gz`.

APT/other (optional)

- No deb/rpm planned for first release; document that Homebrew/script/manual tarball are supported paths.

Script installer

- `scripts/install.sh` must fetch from new release URLs and install buddy + optional alias; verify checksum before install.

Docs

- README quick install: `brew install buddy` (tap path TBD) and `curl -fsSL https://get.buddy.sh | sh` (once URL live).

Testing

- Part of release QA matrix: brew install/uninstall, checksum verification, binary run.
