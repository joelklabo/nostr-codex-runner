# Issue: Mermaid diagram best practices

Context

- Ensure our mermaid diagrams follow GitHub’s recommended practices: <https://docs.github.com/get-started/writing-on-github/working-with-advanced-formatting/creating-diagrams#creating-mermaid-diagrams>
- Current README uses a flowchart; other docs may add more diagrams.

Tasks

- Review README mermaid block for syntax, title, accessibility (labels), and clarity.
- Add guidance to docs/style-guide.md on mermaid usage (themes, labels, avoiding overly dense graphs).
- Optional: add a simple render check (markdownlint rule or lightweight CI script) to catch malformed mermaid.

Acceptance

- README mermaid block complies with GitHub best practices.
- Style guide includes mermaid guidance.
- CI/lint strategy decided (documented) for catching malformed mermaid diagrams.
