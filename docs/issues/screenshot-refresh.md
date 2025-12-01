# Issue: Refresh README/Landing Screenshots (Wizard-first)

Goal

- Replace outdated screenshots with ones that match the current wizard-first quickstart (brew + wizard + preset list) and mock-echo offline note.

Targets

- README quick-start/quick-install section.
- docs/landing.md hero/quickstart snippet (if published to GitHub Pages).

What to capture

- From the rendered README: the "Quick install" code block and the 3-step quick start showing `buddy wizard` and `buddy run <preset>`.
- A second shot with the presets list snippet (claude-dm, copilot-shell, local-llm, mock-echo) if room allows.
- Optional: a narrow/mobile viewport version for landing if we ship it.

How to capture (suggested)

- Use a local markdown renderer or GitHub preview in browser; set OS font smoothing on.
- Width ~1280px for desktop capture; PNG, 2x DPR; keep under 300KB each.
- Name files `assets/screenshot-readme-quickstart.png` and `assets/screenshot-landing-hero.png`.
- Update README/landing to point to the new filenames if embedded.

Acceptance

- Screenshots show wizard-first flow (brew -> wizard -> run) and mention mock-echo as offline/test-only.
- No obsolete copy (old paths or missing BUDDY_CONFIG).
- Files committed under `assets/` and referenced where used.
