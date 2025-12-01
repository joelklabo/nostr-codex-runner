# Wizard Extensibility – Issue 3oa.24

Goal: make it easy to plug new transports/agents/actions/presets into the wizard without rewriting the flow.

Proposed design
- Registry structs exposed from wizard package:
  - `type TransportOption struct { Name, Description string; Prompts []PromptSpec }`
  - `type AgentOption struct { Name, Description string; Prompts []PromptSpec }`
  - `type ActionOption struct { Name, Description string; Prompts []PromptSpec; DefaultEnabled bool }`
  - `type PresetOption struct { Name, Description string; Apply func(*config.Config) }`
- `PromptSpec` describes a question: kind (input/password/select/confirm), label, default, validation func.
- Wizard flow builds options by iterating registries; core flow stays the same.

Flow impact
- Preset selection pre-fills config via `Apply` and skips irrelevant prompts.
- Transports/agents/actions added by registering options in `init()` of their package (or a central registry file).
- Defaults remain the same if no new options are registered.

Testing
- Add table-driven tests feeding custom registries to ensure prompts are invoked and config patched correctly.
- Golden tests for generated configs when registry contents change.

Steps to implement
1) Define registry types and helpers in `internal/wizard/registry.go`.
2) Refactor `wizard.Run` to consume registry options instead of hard-coded choices.
3) Add built-in options for nostr/mock, http/copilot/echo, shell/readfile/writefile, presets (claude-dm, nostr-copilot-shell, local-llm, mock-echo).
4) Update tests to inject minimal registries.
5) Update docs to show how to add new options.

Docs updates
- Add section in `docs/wizard.md` describing registry hook for contributors.
- Mention that new plugins should also add presets where possible for consistency.
