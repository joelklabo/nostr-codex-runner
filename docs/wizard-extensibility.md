# Wizard Extensibility – Issue 3oa.24

Goal: make it easy to plug new transports/agents/actions/presets into the wizard without rewriting the flow.

Proposed design
- Registry structs exist in `internal/wizard/registry.go`:
  - `TransportOption`, `AgentOption`, `ActionOption`, `PresetOption`, plus `PromptSpec`.
- `GetRegistry` returns the current registry; `SetRegistry` overrides it (handy for plugins/tests).
- Wizard flow iterates the registry; selecting a preset simply pulls embedded YAML today.

Flow impact
- Preset selection pre-fills config via `Apply` and skips irrelevant prompts.
- Transports/agents/actions added by registering options in `init()` of their package (or a central registry file).
- Defaults remain the same if no new options are registered.

Testing
- Add table-driven tests feeding custom registries to ensure prompts are invoked and config patched correctly.
- Golden tests for generated configs when registry contents change.

Steps to implement
1) (Done) Registry types + default entries live in `internal/wizard/registry.go`.
2) (Done) `wizard.Run` consumes the registry.
3) (Done) Built-ins: nostr/mock transports; http/copilot/echo agents; shell/readfile/writefile actions; presets: claude-dm, copilot-shell, local-llm, mock-echo.
4) (Done) Tests can inject a custom registry via `SetRegistry`.
5) (TODO) Add contributor snippet in README/CONTRIBUTING on how to add new options.

Docs updates
- Add section in `docs/wizard.md` describing registry hook for contributors.
- Mention that new plugins should also add presets where possible for consistency.
