# Local Dev Setup (Contributors) – Issue 3oa.9

Scope: for contributors. End users should install the `buddy` binary from releases or brew.

Prereqs
- Go ≥1.22
- Codex CLI on PATH if using codex agent tests
- `bd` installed (issue tracking)
- Access to test Nostr relays (for integration flows)

Setup steps
1) Clone:
   ```bash
   git clone https://github.com/joelklabo/buddy.git
   cd buddy
   ```
2) Deps:
   ```bash
   go mod download
   ```
3) Build & run:
   ```bash
   make build      # bin/buddy (+ nostr-buddy symlink)
   make run        # uses config.yaml
   ```
4) Formatting & lint:
   ```bash
   make fmt
   make lint       # go vet ./...
   ```
5) Tests:
   ```bash
   go test ./...
   ```
6) Sample configs:
   - `config.example.yaml` → copy to `config.yaml`
   - `sample-flows/` has preset configs (nostr-copilot, etc.).

Paths & state
- Config: `./config.yaml` for repo work; default runtime path `~/.config/buddy/config.yaml` once rename lands.
- State DB: `state.db` (BoltDB) — excluded by .gitignore; keep test DBs out of git.
- Presets: `assets/presets` (embedded when packaging); user overrides in `~/.config/buddy/presets/`.

Testing notes
- Prefer table-driven tests; place alongside code in `_test.go`.
- For nostr-dependent tests, use mock transport or inject relay URLs via test config.

bd workflow
- One issue per commit; create under epic `buddy-3oa`.
- Commit message should close the issue, e.g., `Some change (closes buddy-3oa.X)`.

After rename
- Module path is `github.com/joelklabo/buddy`.
- Binaries: `buddy` (+ `nostr-buddy` alias).
