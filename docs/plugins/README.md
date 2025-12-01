# Plugin Catalog

The runner is fully pluggable. Pick transports, an agent, and actions in `config.yaml`. These pages summarize what ships today and where to add your own.

- [Transports](transports.md)
- [Agents](agents.md)
- [Actions](actions.md)

Add your plugin under `internal/{transports|agents|actions}/<name>` and register it in `init()` via `registry.MustRegister("<name>", ...)`. Then reference it in `config.yaml` with `type: <name>` plus your custom fields.
