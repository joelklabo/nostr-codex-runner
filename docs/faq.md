# Buddy FAQ / Troubleshooting

## Install

- **`buddy: command not found`** – Ensure `~/bin`, `~/.local/bin`, or `/usr/local/bin` is on PATH. Homebrew: `brew install joelklabo/tap/buddy`. Script: `curl -fsSL https://get.buddy.sh | sh`.

- **Checksum mismatch** – Re-download; verify against `buddy_checksums.txt` from the release. On macOS: `shasum -a 256 buddy_<ver>_macOS_arm64.tar.gz`.

## Running presets

- **`path or preset "<name>" not found`** – Built-ins are `mock-echo`, `claude-dm`, `copilot-shell`, `local-llm`. Check spelling or pass a config path.

- **Copilot errors** – Install GitHub Copilot CLI (`npm i -g @github/copilot`), run `copilot auth login`, then `buddy check copilot-shell`.

- **Claude/OpenAI HTTP errors** – Confirm API key is set in config, and `buddy check claude-dm` shows the endpoint reachable.

## Wizard

- **Wizard aborted: missing dependencies** – Install the listed binaries or rerun with `buddy wizard` and confirm when prompted; or run `buddy check <preset>` first.

- **Overwrite prompt** – Wizard won’t clobber an existing config without confirmation. Pass a new path to write elsewhere.

## Config and paths

- Search order: positional/`-config`, then `./config.yaml`, then `~/.config/buddy/config.yaml`.

- Preset overrides: place custom YAML at `~/.config/buddy/presets/<name>.yaml`.

- Logs: default stdout; set `logging.file` to write to `~/.buddy/runner.log`.

- State DB: default `~/.buddy/state.db` unless overridden.

## Relays / transport

- **Connection refused/timeouts** – Try different relays (`wss://relay.damus.io`, `wss://nos.lol`), or use `mock-echo` to validate the pipeline offline.

- **`missing allowed_pubkeys`** – Ensure `runner.allowed_pubkeys` has at least one hex npub; wizard will prompt for it.

## Actions / safety

- Shell is high risk. Keep `allowed_pubkeys` tight, and prefer `mock-echo` or `claude-dm` until ready.

- To disable shell in a preset override, remove the `shell` action or set stricter allowlists.

## Metrics / health

- Enable health endpoint with `-health-listen 127.0.0.1:8081`; metrics via `-metrics-listen 127.0.0.1:9090`.

## Windows

- Not yet supported; use WSL with the Linux binary.

## Still stuck?

- Run `buddy check <preset>` and share the output (redact secrets).

- Open an issue with OS/arch, buddy version (`buddy version`), and the failing command/output.
