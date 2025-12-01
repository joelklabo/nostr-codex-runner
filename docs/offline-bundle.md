# Offline Bundle Considerations – Issue 3oa.34

Goal: buddy should run with built-in presets and example configs even without network access (aside from chosen transport/agent).

Plan

- Embed default presets and `config.example.yaml` into the binary (via go:embed) so `buddy presets` works offline.

- Wizard: allow `mock` transport path that requires no network; default relays still offered but user can choose mock.

- Agent: include `echo` preset for offline smoke; local-LLM preset can point to localhost.

- Install script: avoid network after download; no extra package installs.

- Run behavior: if network-dependent transport fails, suggest `buddy run mock-echo` for offline verification.

Testing

- Airplane-mode test in QA matrix: `buddy presets`, `buddy run mock-echo`, `buddy wizard --dry-run` should succeed without network.
