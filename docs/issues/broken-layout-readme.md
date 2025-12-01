# Issue: README layout glitches (per screenshot)

Context

- Screenshot R5vguv.png shows broken layout (Quick install / Quick start block rendering).
- Need to reproduce on GitHub-rendered README and fix spacing/Markdown formatting.

Tasks

- Reproduce by viewing README on GitHub; note exact breakage.
- Fix Markdown spacing/fences so lists/code blocks render correctly.
- Run markdownlint locally and ensure CI stays green.

Acceptance

- README renders correctly on GitHub (no truncated code blocks or list glitches).
- Markdownlint passes.
