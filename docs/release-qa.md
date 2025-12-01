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
5) `buddy run copilot-shell` (or claude-dm) starts with sample config/preset (if secrets available, otherwise skip)
6) `buddy presets` lists built-ins
7) `buddy help` shows commands
8) `buddy check mock-echo -json` prints JSON with status fields
9) `buddy presets claude-dm --yaml` respects overrides if `~/.config/buddy/presets/claude-dm.yaml` exists
10) Homebrew tap check: `brew tap joelklabo/tap && brew info buddy && brew install joelklabo/tap/buddy && buddy --version` (CI job `homebrew-check.yml`).

Brew checklist (macOS)

- `brew install joelklabo/tap/buddy`

- `buddy --version`

- `brew uninstall buddy`

Alias checks

- If alias packaged: `nostr-buddy --version` maps to same build; ensure warning about legacy binary name not printed when using alias.

Logging/metrics

- Optional: run with `-health-listen 127.0.0.1:8081 -metrics-listen 127.0.0.1:9090`; curl `/health` and `/metrics` (expect 200 + Prometheus text format).

Man page

- `make man` succeeds on release branch.

- Extract archive and ensure `share/man/man1/buddy.1` is present; `man -M <extract>/share/man buddy` renders without errors (macOS/Linux).

Fail-fast criteria

- Version mismatch, missing preset list, wizard failure, or checksum mismatch abort release.

Notes

- Use public relays for nostr tests only if keys available; otherwise use mock transport.

- Document results in release PR checklist.
