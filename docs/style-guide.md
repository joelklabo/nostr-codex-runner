# Documentation Style Guide

Tone: concise, friendly, actionable. Prefer second person (“you”), avoid filler, keep sections scannable.

Headings: Title Case, no more than two levels deep per page. Start with a one-line summary under H1 when helpful.

Lists: use `-` bullets; keep lines short; avoid nested lists unless necessary.

Code fences: use language hints (`bash`, `yaml`, `go`). For commands, show the exact line to paste; keep outputs short.

Examples: prefer runnable snippets (`buddy run mock-echo`), include expected output when clarifying success.

Links: relative paths within the repo (e.g., `docs/faq.md`), no raw URLs in prose—use link text.

Tables: only when comparing options; keep under ~80 characters per cell.

Admonitions: keep to short paragraphs; favor inline notes over long warnings.

Accessibility: spell out emoji meaning in text when used for status (e.g., “✅ ok”).

Templates:

- How-to: Goal → Steps (numbered) → Verify → Troubleshooting.

- Concept: What/Why → Key terms → How it fits → Links to how-tos.

- Reference: Field/name → Type/default → Example → Notes.

Mermaid diagrams:

- Use fenced blocks with language `mermaid`; for styling include an init line like `%%{init: {'theme': 'neutral'}}%%`.

- Prefer LR direction to avoid tall diagrams in README.

- Keep labels concise; avoid HTML unless necessary for line breaks; follow GitHub guidance (<https://docs.github.com/get-started/writing-on-github/working-with-advanced-formatting/creating-diagrams#creating-mermaid-diagrams>).

- Preview on GitHub before merging; broken diagrams should be fixed or removed.
