# Help / Man Page Plan – Issue 3oa.35

Goals
- Concise `buddy help` output matching CLI map (run/check/wizard/presets/init-config/version).
- Optional `man buddy` generated during release packaging.

CLI help content (target)
```
buddy run <preset|config>      start the runner from a preset or YAML config
buddy wizard [config-path]     guided setup; writes a config (supports dry-run)
buddy presets [name]           list built-in presets or show details
buddy check <preset|config>    verify dependencies (use --json for machine output)
buddy init-config [path]       write example config if missing
buddy version                  show version info
buddy help [cmd]               show this help
```

Man page generation
- Use `go generate` + simple template under `docs/man/buddy.1.md` -> rendered via `ronn` or `pandoc` in CI, or use `go-md2man` during goreleaser.
- Include synopsis, description, commands, environment (`BUDDY_CONFIG`, legacy NCR warning), files (config path, presets path), exit codes, examples.

Packaging
- Ship man page in tar.gz under `share/man/man1/buddy.1`.
- Homebrew formula installs the man page when available.

Testing
- Unit test to ensure `help` prints command list without panic.
- Manual spot-check in release QA: `man ./buddy` after install if manpath set.
