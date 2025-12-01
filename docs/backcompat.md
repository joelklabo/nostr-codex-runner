# Backward Compatibility & Migration – Issues 3oa.32 / 3oa.43

Legacy invocation
- Old binary: `nostr-codex-runner run -config config.yaml`
- Legacy env: `NCR_CONFIG` pointed to config path.
- Legacy config path: `~/.config/nostr-codex-runner/config.yaml`

New path
- Binary names will become `buddy` (canonical) and alias `nostr-buddy` to avoid collisions.
- Env: `BUDDY_CONFIG` (preferred), still honors `NCR_CONFIG` with a warning.
- Default config path: `~/.config/buddy/config.yaml`.

Runtime shims (current)
- CLI accepts `NCR_CONFIG` and legacy config locations, but prints warnings to stderr:
  - Using `NCR_CONFIG` is deprecated; use `BUDDY_CONFIG`.
  - Legacy config dir notice.
  - Binary name notice if `nostr-codex-runner` is invoked.

Migration steps for users
1) Move config to `~/.config/buddy/config.yaml` (or pass via `-config`).
2) Switch env var to `BUDDY_CONFIG`.
3) When the buddy binary is published, install it and use `buddy run ...` (alias `nostr-buddy` if collisions matter).

Timeline
- One release will ship with both names and warnings.
- Subsequent release may drop `NCR_CONFIG` and legacy paths; keep alias for one more cycle.
