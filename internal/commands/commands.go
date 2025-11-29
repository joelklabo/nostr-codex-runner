package commands

import (
    "strings"
)

// Command represents a parsed user instruction carried over Nostr.
type Command struct {
    Name   string // run|new|reset|use|status|help
    Args   string // remaining text after the command keyword
    Raw    string // original user message
}

// Parse inspects the incoming plaintext message and extracts a command.
// Supported forms:
//   "/new" or "new"                -> start a new session (Args optional)
//   "/use <session-id>"            -> switch to an existing session
//   "/status"                      -> show active session info
//   "/help"                        -> usage help
//   Anything else                  -> run prompt in the active/new session
func Parse(msg string) Command {
    trimmed := strings.TrimSpace(msg)
    lower := strings.ToLower(trimmed)

    switch {
    case strings.HasPrefix(lower, "/new"):
        return Command{Name: "new", Args: strings.TrimSpace(trimmed[4:]), Raw: msg}
    case strings.HasPrefix(lower, "new"):
        return Command{Name: "new", Args: strings.TrimSpace(trimmed[3:]), Raw: msg}
    case strings.HasPrefix(lower, "/reset"):
        return Command{Name: "new", Args: strings.TrimSpace(trimmed[6:]), Raw: msg}
    case strings.HasPrefix(lower, "reset"):
        return Command{Name: "new", Args: strings.TrimSpace(trimmed[5:]), Raw: msg}
    case strings.HasPrefix(lower, "/use"):
        return Command{Name: "use", Args: strings.TrimSpace(trimmed[4:]), Raw: msg}
    case strings.HasPrefix(lower, "use"):
        return Command{Name: "use", Args: strings.TrimSpace(trimmed[3:]), Raw: msg}
    case strings.HasPrefix(lower, "/status"):
        return Command{Name: "status", Raw: msg}
    case strings.HasPrefix(lower, "status"):
        return Command{Name: "status", Raw: msg}
    case strings.HasPrefix(lower, "/help"):
        return Command{Name: "help", Raw: msg}
    case strings.HasPrefix(lower, "help"):
        return Command{Name: "help", Raw: msg}
    default:
        return Command{Name: "run", Args: trimmed, Raw: msg}
    }
}
