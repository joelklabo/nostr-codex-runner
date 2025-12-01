# Windows Support Decision – Issue 3oa.36

Current stance: not officially supported in first buddy release.

Reasons

- Nostr relay tooling and deps tested only on macOS/Linux.
- Copilot/Codex CLI prerequisites assumed POSIX shell.
- No Windows CI runner configured; systemd/launchd recipes non-portable.

What to document

- README/FAQ: "Windows not yet supported; use WSL2 or Linux/macOS."
- Provide minimal guidance for WSL2 with Ubuntu: install Go, use Linux binaries, ensure relays reachable.

Future path (optional)

- Add GitHub Actions Windows build matrix and smoke test (`buddy presets`, `buddy wizard --dry-run`).
- Consider Scoop/winget package once binary works natively.
