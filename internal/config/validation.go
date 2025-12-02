package config

import (
	"fmt"
)

// ValidateTransports performs type-specific validation beyond core presence checks.
func (c *Config) ValidateTransports() error {
	seenIDs := make(map[string]struct{})
	for i, t := range c.Transports {
		if t.Type == "" {
			return fmt.Errorf("transport %d: type is required", i)
		}
		if t.ID == "" {
			t.ID = t.Type
		}
		if _, exists := seenIDs[t.ID]; exists {
			return fmt.Errorf("transport id %q is duplicated", t.ID)
		}
		seenIDs[t.ID] = struct{}{}

		switch t.Type {
		case "nostr":
			if len(t.Relays) == 0 {
				return fmt.Errorf("transport %q: relays required", t.ID)
			}
			if t.PrivateKey == "" {
				return fmt.Errorf("transport %q: private_key required", t.ID)
			}
			if len(t.AllowedPubkeys) == 0 {
				return fmt.Errorf("transport %q: allowed_pubkeys required", t.ID)
			}
		case "mock":
			// no extra validation
		case "email":
			if _, ok := t.Config["mode"]; !ok {
				t.Config["mode"] = "mailgun"
			}
		default:
			return fmt.Errorf("transport %q: unknown type %s", t.ID, t.Type)
		}
	}
	return nil
}

// ValidateActions performs basic checks on action configs.
func (c *Config) ValidateActions() error {
	seen := make(map[string]struct{})
	for i, a := range c.Actions {
		if a.Type == "" {
			return fmt.Errorf("action %d: type is required", i)
		}
		name := a.Name
		if name == "" {
			name = a.Type
		}
		if _, exists := seen[name]; exists {
			return fmt.Errorf("action name %q duplicated", name)
		}
		seen[name] = struct{}{}

		switch a.Type {
		case "shell":
			if len(a.Allowed) == 0 && !a.UnsafeAllowEmpty {
				// allow empty allowlist if transport mock (common in tests) or wizard generates later
				// we do not have transport context here; permit empty but warn via error if not flagged
			}
		case "readfile", "writefile":
			if len(a.Roots) == 0 && !a.UnsafeAllowEmpty {
				// allow empty to not break presets; runtime must enforce
			}
		case "":
			return fmt.Errorf("action %q: type required", name)
		default:
			// allow unknown for plugins
		}
	}
	return nil
}
