# Benchmarks & Performance Notes – Issue 3oa.13

What to expect (rough order-of-magnitude)
- Nostr relay latency: 100–500 ms typical per hop; DM roundtrip depends on relays used.
- Agent latency: Codex/Claude/OpenAI: 1–5s per prompt; local LLM varies with model/hardware.
- Action latency: shell/readfile usually <100ms unless command is heavy.

Knobs that affect performance
- `agent.config.timeout_seconds`: cap slow model calls.
- `actions.shell.timeout_seconds` and `max_output`: prevent long-running commands and large payloads.
- `runner.max_reply_chars`: truncation helps slow transports.
- Relay choice and count: fewer relays may reduce fan-out delay; prefer reliable relays.

Profiling tips
- Enable debug logging temporarily to see timing per request.
- Use `pprof` on the binary when running locally: set `GODEBUG`/`runtime/pprof` hooks as needed (not wired by default).
- For action-heavy flows, profile the external command, not the runner.

Load considerations
- The runner is single-binary; concurrency comes from goroutines per transport and action. Monitor CPU/memory via `top` or Prom metrics if enabled.
- Storage is BoltDB; heavy parallel writes may contend; keep action results small.

Recommendations
- Set conservative timeouts for shell actions in production.
- Use local model preset for offline/low-latency scenarios.
- Keep logs at info; switch to debug only when diagnosing delays.
