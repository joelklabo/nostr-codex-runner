# Agents

Built-ins:

| Type          | Package                           | Notes / config keys                                   |
|---------------|-----------------------------------|-------------------------------------------------------|
| `codexcli`    | `internal/agents/codexcli`        | Default; CLI flags via `agent.config.*`.              |
| `copilotcli`  | `internal/agents/copilotcli`      | Uses `copilot` (github/copilot-cli).                  |
| `echo`        | `internal/agents/echo`            | No-op echo for testing.                               |
| `http`        | `internal/agents/http`            | Stub for OpenAI/Claude-style APIs.                    |

Config fields are read from `agent.config` (legacy alias: `agent.codex`). Common fields:
- `binary` (string)
- `working_dir` (string)
- `timeout_seconds` (int)
- `extra_args` (list)

Add an agent:
1) Create `internal/agents/<name>/` implementing `core.Agent`.
2) Register with `agent.MustRegister("<name>", New)` in `init()`.
3) Define a config struct for your fields and parse `agent.config` (or a namespaced map if you need more).
4) Document a sample under `docs/recipes/`.
