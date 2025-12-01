# Module & Binary Rename Plan – Issue 3oa.39

Current module: `module github.com/joelklabo/buddy`
Current binary: `buddy`

Steps
1) After repo rename (3oa.38), update `go.mod` module path and run `go mod tidy`.
2) Update import paths across repo (`github.com/joelklabo/nostr-codex-runner` → `github.com/joelklabo/buddy`).
3) Tests/CI:
   - Update workflow caches and binary names (artifacts in CI/release).
   - Ensure `go test ./...` passes with new module path; refresh coverage badges.
4) External references:
   - Update package docs badge, go reference badge, README commands, systemd template, recipes, sample configs.
5) Validation steps:
   - Fresh clone after rename; `go test ./...`.
   - `make build` produces `bin/buddy` and `bin/nostr-buddy` symlink.
   - `buddy version` shows new module path.

Risks
- Downstream imports break: mitigate with replace directive temporarily and clear release notes.
- Users with old binary names: provide shim and warning.
