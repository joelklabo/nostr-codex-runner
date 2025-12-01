# Transports

Built-ins:

| Type    | Package                                  | Notes / config keys                         |
|---------|------------------------------------------|---------------------------------------------|
| `nostr` | `internal/transports/nostr`              | `relays`, `private_key`, `allowed_pubkeys`  |
| `mock`  | `internal/transports/mock`               | For tests; echoes messages in-process.      |
| `slack` | `internal/transports/slack` (stub)       | Scaffold only; fill in token/app config.    |

To add a transport:
1) Create `internal/transports/<name>/` with a type implementing `core.Transport`.
2) Call `transport.MustRegister("<name>", New)` in `init()`.
3) Accept your config fields in a `Config` struct and validate them.
4) Document a sample block in `config.example.yaml` or a recipe under `docs/recipes/`.
