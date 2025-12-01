# Buddy Example Use Cases (argument-only)

Fast, copy-paste flows that assume you installed the `buddy` binary and have keys ready. Easiest path: run `buddy wizard` once to generate `~/.config/buddy/config.yaml`, then use the presets below. Commands avoid flags; use positional arguments and presets.

## 1) Nostr DM to Claude/OpenAI (hosted model)

- Start the preset:

  ```bash
  buddy run claude-dm
  ```

- What happens: buddy loads the `claude-dm` preset, connects to your configured relays, and listens for DMs from allowed pubkeys. The preset uses the Claude/OpenAI HTTP agent and defaults to the nostr transport.

- Try it: from an allowed pubkey, send

  ```text
  /new
  Write a haiku about tunnels.
  ```

- Expected: terminal log line showing relay subscriptions; DM reply with `session: <id>` and model output. Stop with Ctrl+C.

## 2) Local/offline model flow

- Start the preset (example using local LLM agent):

  ```bash
  buddy run local-llm
  ```

- What happens: loads the local LLM preset, points the agent at your local endpoint/binary (see preset notes), keeps traffic on your machine. Nostr is the default transport; to stay fully offline, drop a preset override at `~/.config/buddy/presets/local-llm.yaml` that sets `type: mock` for transport.

- Try it: send

  ```text
  /new Summarize the last git commit in 2 sentences.
  ```

- Expected: response from your local model; good for air-gapped or privacy-sensitive use.

## 3) Custom action trigger / shell copilot

- Start the preset:

  ```bash
  buddy run copilot-shell
  ```

- What happens: combines nostr transport + Copilot agent + `shell` action. The runner will honor allowlists and action timeouts.

- Try it:

  ```text
  /shell ls -la
  /new Write a bash script that tails logs and filters ERROR.
  ```

- Expected: shell output truncated per config; model replies with the script. Keep this preset to trusted operators only.

## Tips

- Config search order: CLI arg path > `BUDDY_CONFIG` > `./config.yaml` > `~/.config/buddy/config.yaml`.

- Presets are embedded but are overridden by `~/.config/buddy/presets/<name>.yaml`, then `./presets/<name>.yaml` if present.

- If another `buddy` binary is on your PATH (e.g., Buddy.Works), invoke via the alias you choose (e.g., `nostr-buddy`).

- Recipe: see `docs/recipes/local-llm-mock.md` for a fully offline local-llm setup.
