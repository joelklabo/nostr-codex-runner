# Help/Man Implementation Plan (execution)

Tasks
1) Add `help` subcommands and `--help` flags to match CLI map; ensure `buddy help [cmd]` works.
2) Generate man page using `go-md2man` during release:
   - Add `docs/man/buddy.1.md` (markdown source) with synopsis, commands, env, files, examples, exit codes.
   - Add a `make man` target that outputs `docs/man/buddy.1` using `go-md2man` (install if missing).
   - Wire goreleaser to include `docs/man/buddy.1` in archives under `share/man/man1/` and `brew` formula `man1.install`.
3) Update `buddy help` to print concise usage plus command-specific help when arg provided.
4) Tests: ensure `help` returns exit 0 and contains command list; add golden tests for `help run` and `help wizard`.
5) Docs: mention man page availability in README quick links.

Notes
- Keep output short; no ANSI.
- `buddy version` should remain simple; `--json` flag can be added later if needed.
