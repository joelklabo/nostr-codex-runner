# Changelog Policy – Issue 3oa.11

Proposal
- Maintain `CHANGELOG.md` in Keep a Changelog style.
- New entry per release tag; group by Added/Changed/Fixed/Removed/Docs.
- Date format: YYYY-MM-DD.

Process
- During release PR, update `CHANGELOG.md` with entries for the upcoming version.
- Include: headline features (wizard, presets), breaking changes (binary rename), deprecations (old env vars), bug fixes.
- Link to compare view: `https://github.com/joelklabo/buddy/compare/vX.Y.Z...vA.B.C` (update after rename).

Versioning
- Semantic versioning (MAJOR.MINOR.PATCH).
- Breaking changes (e.g., binary rename) bump MAJOR.

Tooling
- Optional helper: `git cliff` or simple manual edits; keep short.
- Ensure goreleaser release notes point to the changelog section.

Backfill
- Add entries for last released version under old name; note rename in the first buddy release.
