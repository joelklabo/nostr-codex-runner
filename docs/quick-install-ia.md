# Quick Install Path (binary-first) – Outline

Goal: 2-minute path to download buddy, verify, run the wizard or a preset. No git clone.

## Flow to document

1. Choose install method

   - Homebrew: `brew install joelklabo/tap/buddy`.
   - Script: `curl -fsSL https://get.buddy.sh | sh` (installs to ~/.local/bin or /usr/local/bin based on perms; verifies checksum).
   - Manual: download tarball from Releases, `chmod +x buddy`, move to PATH.
   - Collision note: if another `buddy` exists (Buddy.Works, Livebud), add our binary earlier on PATH or create your own alias (e.g., `bud`).

2. Verify install

   - `buddy --version` (shows version + git commit).
   - `buddy help` (short usage table).
   - For brew installs, `make brew-check` replicates the CI install/test locally (macOS).

3. Prepare config via wizard (fastest default)

   - `buddy wizard` → answers: transport (nostr default), relays, private key input (hidden), allowed pubkeys, agent choice (Claude/OpenAI/local), action selection. Wizard writes `~/.config/buddy/config.yaml` (or provided path) and prints next command.

4. Run a preset (no edits)

   - `buddy run copilot-shell` (uses built-in preset; prompts for necessary secrets if missing; points at default relays).
   - Alternative: `buddy run claude-dm` or `buddy run local-llm` (if offline/local model configured).

5. 2-minute smoke test

   - Show expected output after start: logs line with listening relay/preset; instructions: "DM me from allowed pubkey: /new hello" for nostr, or use mock preset for offline (`buddy run mock-echo` runs without relays; inject via tests, not DMs).
   - Success criteria: receives first response; exits on Ctrl+C.

6. Where files live

   - Binary path, config search order: argv path > `BUDDY_CONFIG` env > cwd config.yaml > `~/.config/buddy/config.yaml`.
   - State DB: `~/.buddy/state.db`. Logs: stdout by default; optional `~/.buddy/runner.log`.
   - Presets directory (embedded / assets/presets) and user overrides (`~/.config/buddy/presets` if present).

7. Prerequisites

   - OS/arch supported: macOS (arm64/amd64), Linux (arm64/amd64).
   - Agent-specific deps: Copilot CLI or Claude/OpenAI HTTP keys; local LLM requirements if chosen.

8. Uninstall

   - Homebrew: `brew uninstall buddy`.
   - Script/manual: remove binary + config folder; note BoltDB state path.

## Notes

- Add checksum sample snippet using `buddy_<version>_<os>_<arch>.tar.gz` and `buddy_checksums.txt`.
