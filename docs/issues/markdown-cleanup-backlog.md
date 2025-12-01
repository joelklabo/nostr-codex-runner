# Issue: Markdown Cleanup Backlog

Context

- markdownlint is now scoped to README + landing + key issue docs to keep CI green.
- Legacy docs/templates violate many rules (blank lines, fenced code spacing, ordered list prefixes).

Tasks

- Incrementally bring remaining Markdown files into compliance or update lint config with explicit rationale for exclusions.
- Prioritize: CONTRIBUTING.md, CHANGELOG.md, CODE_OF_CONDUCT.md, SECURITY.md, docs/*.md, .github templates.
- Consider auto-format (markdownlint --fix) where safe.

Acceptance

- markdownlint scope expanded (or exclusions documented) with zero violations.
- No critical docs (.github templates, CONTRIBUTING) are excluded without a note in `.markdownlint.yaml`.
