# Actions

Built-ins:

| Type         | Package                            | Notes / config keys                               |
|--------------|------------------------------------|---------------------------------------------------|
| `shell`      | `internal/actions/shell`           | `workdir`, `allowed[]` prefixes, `timeout_seconds`, `max_output`. |
| `readfile`   | `internal/actions/fs` (readfile)   | `roots[]`, `max_bytes`.                           |
| `writefile`  | `internal/actions/fs` (writefile)  | `roots[]`, `allow_write`, `max_bytes`.            |

Defaults: if you omit `actions`, a `shell` action is auto-added with safe defaults so `/raw` works out of the box.

Add an action:
1) Create `internal/actions/<name>/` implementing `core.Action`.
2) Register with `action.MustRegister("<name>", New)` in `init()`.
3) Define a config struct for your fields and parse from `actions[].<field>`.
4) Document a sample under `docs/recipes/`.
