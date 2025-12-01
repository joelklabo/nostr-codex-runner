# Buddy Docs, Wizard, and CLI Productization Epic

Goal: ship a binary-first "buddy" CLI with a fast install, guided wizard, presets, and clear docs for users and contributors.

## Issue list (Todo unless noted)

### Docs & Onboarding
1. P0 README information architecture (binary-first; quickstart with `buddy wizard` and presets). ✅
2. P0 Quick install path (curl/Homebrew; 2-minute smoke using a preset; no git clone).
3. P0 Example use cases page (argument-only UX; presets + wizard outputs). ✅ draft
4. P0 Docs landing/index (User vs Contributor entry; links to wizard/presets). ✅ draft
5. P1 Architecture diagram (include CLI front door and preset loader).
6. P1 Extend/commands guide (register new CLI subcommands/presets).
7. P1 Config reference (search order, preset schema, defaults).
8. P1 Contributing refresh (CLI-first stance; existing standards).
9. P1 Local dev setup (contributors only).
10. P1 Docs style guide/templates.
11. P2 Changelog policy.
12. P2 FAQ/Troubleshooting (install/preset/wizard).
13. P2 Benchmark notes.
14. P2 Security/secrets (wizard + CLI logging).
15. P3 Website/landing polish.
16. P3 Precedent research (great onboarding examples).

### Wizard Track
17. P0 Wizard concept (goals, scope, success metrics). ✅
18. P0 Wizard IA/script (questions → config + preset choice). ✅
19. P1 Wizard UX prototype (survey/promptui/bubbletea vs bufio).
20. P1 Wizard implementation (`buddy wizard`; writes config; optional smoke run). ✅
21. P1 Wizard tests (branching, goldens, stdin sim). ✅ initial
22. P1 Wizard docs (README blurb + page; clip). ✅ initial
23. P2 Wizard telemetry/safety (no secret logging; dry-run).
24. P2 Wizard extensibility (registry for new actions/providers).
25. P3 Wizard polish (presets, color toggle, retries).

### CLI Productization
26. P0 CLI spec/map (`buddy wizard`, `buddy run <preset|config.yaml>`, `buddy presets`, `buddy help`; arguments over flags). ✅ v1
27. P0 Preset library (ship built-ins: `copilot-shell`, `claude-dm`, `local-llm`; assets/presets). ✅
28. P0 Packaging & releases (goreleaser, checksums, Homebrew tap).
29. P1 Install script (curl | sh; checksum; /usr/local/bin or ~/.local/bin).
30. P1 CLI UX copy/errors (friendly, masked secrets, exit codes).
31. P1 Config search precedence (arg path > cwd config.yaml > ~/.config/buddy/config.yaml; env opt-in).
33. P1 Release QA matrix (macOS/Linux/arm64; presets + wizard).
34. P2 Offline bundle (embed default presets/assets; graceful no-network). ✅ presets embedded; mock-echo default in wizard.
35. P2 Help/man page generation. 🚧
36. P3 Windows support decision (scoop/winget or "not yet").

### Name & Migration
37. P0 Name locked: "buddy".
38. P0 Repo rename plan (`nostr-codex-runner` → `buddy` repo; redirects, CI updates). ✅ done
39. P0 Module/binary rename (Go module path, imports, main package; build `buddy`). ✅ done
40. P1 Package manager updates (Homebrew formula rename, release artifacts).
41. P1 Brand migration in docs (README, docs index, wizard copy, badges). ✅ mostly done
42. P1 CLI help/man text with new name. ✅
44. P2 Domain/SEO check (distinct from Buddy.Works; optional microsite).

## Notes
- One issue per commit; run tests after each commit.
- Default collision mitigation: ship `buddy` plus optional alias (`bud` off by default).
- Rename done; legacy references now only for backcompat notes.
