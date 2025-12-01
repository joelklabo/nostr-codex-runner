% buddy(1) buddy 2025-12-01
% buddy CLI
% buddy manual

# NAME
buddy - pluggable transport → agent → actions runner with presets and wizard

# SYNOPSIS
**buddy run** <preset|config> [ -config path ] [ -health-listen addr ] [ -metrics-listen addr ] [ -skip-check ]

**buddy wizard** [config-path]

**buddy presets** [name]

**buddy check** <preset|config> [ -config path ] [ -json ]

**buddy init-config** [path]

**buddy version**

**buddy help** [command]

# DESCRIPTION
buddy is a single binary that routes inbound messages (e.g., Nostr DMs) to an AI agent and optional host actions. It ships with presets and a guided wizard so new users can get a working config without editing YAML.

# COMMANDS
**run**  
Start the runner using a preset name or a YAML config path. Flags: -config, -health-listen, -metrics-listen, -skip-check (skip dependency preflight).

**wizard**  
Interactive setup. Prompts for transport/relays/keys, agent choice, actions, and writes a config (supports dry-run).

**presets**  
List built-in presets or print one as YAML when a name is provided.

**check**  
Verify dependencies declared by config or preset. Flags: -config, -json.

**init-config**  
Write the bundled example config to ./config.yaml (or the provided path) if missing.

**version**  
Print version info.

**help**  
Show summary or command-specific help.

# ENVIRONMENT
**BUDDY_CONFIG** — default config path when -config is not provided.

# FILES
`~/.config/buddy/config.yaml` — default config path.  
`~/.config/buddy/presets/` — user preset overrides.  
`~/.buddy/state.db` — BoltDB state (session mapping, cursors).

# EXIT STATUS
0 on success, non-zero on errors.

# EXAMPLES
Run with a preset:  
`buddy run mock-echo`

Generate config, then run:  
`buddy wizard`  
`buddy run -config ~/.config/buddy/config.yaml`

List presets:  
`buddy presets`

# SEE ALSO
README, docs/wizard.md, docs/presets.md
