# Name Decision: buddy (alias opt-in) – Issue 3oa.37

Decision
- Canonical binary/package name: `buddy`.
- Aliases:
  - `nostr-buddy` shipped as a symlink/alt binary to avoid PATH collisions.
  - `bud` remains **opt-in** only (disabled by default) due to conflicts with Livebud/bud.js.

Collision notes
- Existing `buddy` CLI from Buddy.Works CI/CD; may already be on user PATH.
- Existing `bud` binaries in Go/JS ecosystems.
- Docs/installer must warn about potential collision and show how to prefer `nostr-buddy`.

Installer/packaging implications
- Homebrew formula provides `buddy`; optionally a `nostr-buddy` alias via `bin.install_symlink`.
- Install script should detect existing `buddy` in PATH and offer to install as `nostr-buddy` instead.
- Release assets should include checksums for both names if dual binaries are shipped; otherwise provide a post-install symlink step.

Docs updates
- README Quickstart: primary commands use `buddy`; footnote about collisions and alias.
- FAQ: "I already have buddy on PATH" → use `nostr-buddy` or adjust PATH order.
- Wizard copy: mention alias briefly in welcome/collision note.

Open items
- Final decision whether to publish separate `bud` asset or only allow local symlink (`ln -s buddy bud`). Currently leaning to documentation-only, no shipped `bud` artifact.
- Verify Homebrew tap policy on duplicate binary names; ensure caveat text if needed.
