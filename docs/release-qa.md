# Release QA Matrix – Issue 3oa.33

Targets
- macOS arm64 (M-series)
- macOS amd64 (Intel)
- Linux amd64 (Ubuntu/Debian)
- Linux arm64 (Raspberry Pi 4/EC2 aarch64)

Artifacts under test
- buddy binary (and nostr-buddy alias if shipped)
- tar.gz + checksums from GitHub Releases
- Homebrew formula install/uninstall

Smoke checklist (per platform)
1) Download & verify checksum for the correct archive
2) `./buddy --version` matches tag
3) `buddy run mock-echo` starts and logs (Ctrl+C to stop)
4) `buddy wizard --dry-run` completes without writing
5) `buddy run nostr-copilot-shell` (or claude-dm) starts with sample config/preset (if secrets available, otherwise skip)
6) `buddy presets` lists built-ins
7) `buddy help` shows commands

Brew checklist (macOS)
- `brew install joelklabo/tap/buddy`
- `buddy --version`
- `brew uninstall buddy`

Alias checks
- If alias packaged: `nostr-buddy --version` maps to same build; ensure warning about legacy binary name not printed when using alias.

Logging/metrics
- Optional: run with `-health-listen 127.0.0.1:8081 -metrics-listen 127.0.0.1:9090` and curl `/health`.

Fail-fast criteria
- Version mismatch, missing preset list, wizard failure, or checksum mismatch abort release.

Notes
- Use public relays for nostr tests only if keys available; otherwise use mock transport.
- Document results in release PR checklist.
