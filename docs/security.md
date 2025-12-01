# Security & Secrets

## Principles

- Keep secrets local: configs live under `~/.config/buddy`, never committed.

- Mask inputs: wizard uses password prompts for private keys/API keys and prints no secrets.

- Log hygiene: default logging to stdout; avoid including secrets in prompts or configs; optional file log defaults to `~/.buddy/runner.log` with 0600 perms.

## What not to log

- Nostr private keys, API keys, bearer tokens, relay DMs, decrypted content.

- Shell command outputs that may include secrets—prefer redaction if piping through agents.

## Runner hardening

- Restrict `runner.allowed_pubkeys` to trusted operators.

- Prefer private relays for production; avoid relying on open relays for sensitive workloads.

- Use mock transport for local/offline testing to avoid network egress.

- Keep `actions.shell` disabled unless required; scope allowlists and max_output.

- Health/metrics listeners should bind to localhost unless explicitly exposed.

## Dependency checks

- `buddy check <preset>` and run preflight surface missing binaries/ports/relays before startup; fix those before enabling shell or external agents.

## Sandbox & blast radius

- Keep workdirs limited; if using shell action, run the binary in a container/VM where possible.

- Set `actions.*.roots` narrowly for read/write actions.

## Secret storage

- Config values are plain YAML; rely on OS file perms (0600) and keep in user home. No keychain/KMS integration yet.

- State DB (`~/.buddy/state.db`) may contain session transcripts; restrict access and avoid multi-user sharing.

## Release integrity

- Releases include SHA256 checksums and optional cosign signatures (when configured). Verify downloads before running.

## Red-team checklist (quick)

- Are relays private? If not, assume DMs can be observed.

- Are allowed_pubkeys limited? If not, anyone can DM and trigger actions.

- Is shell enabled? If yes, is it strictly allowlisted and monitored?

- Are secrets absent from logs/metrics?

- Are health/metrics endpoints bound to localhost?
