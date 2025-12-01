# FAQ / Troubleshooting – Issue 3oa.12

**Install picks wrong buddy (collision)**
- Symptom: running `buddy` opens Buddy.Works CLI. Fix: install alias `nostr-buddy` or put our binary earlier in PATH.

**Config not found**
- Symptom: "load config" error. Fix: pass path explicitly `buddy run path/to/config.yaml` or place config in cwd; default search order will be argv > ./config.yaml > ~/.config/buddy/config.yaml.

**Relay auth/connection issues**
- Check relay URLs in config; ensure network reachable. Try a known public relay (e.g., wss://relay.damus.io). Run with `logging.level: debug` to see subscription errors.

**BoltDB locked**
- Symptom: `timeout acquiring file lock` on state.db. Fix: stop other running instance or remove stale lock by removing the DB if data can be discarded.

**Missing keys/secrets**
- Wizard/presets need private key or API keys. Ensure the fields are set; secrets are not loaded from env unless explicitly configured.

**Shell action denied**
- Symptom: `/shell ...` returns allowlist/timeout error. Fix: enable `shell` action in config with `allowed_pubkeys`/roots/timeouts; keep it off for untrusted users.

**Output truncated**
- If replies are cut, increase `runner.max_reply_chars` or `actions.shell.max_output`.

**Copilot CLI errors**
- Ensure `copilot` binary is on PATH and authenticated (`copilot auth login`).

**Local model not responding**
- Confirm local endpoint is running; check URL in `local-llm` preset; raise timeout if slow.
