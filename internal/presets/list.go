package presets

import (
	"fmt"
	"os"
	"path/filepath"
)

// List returns preset names and descriptions.
func List() map[string]string {
	return map[string]string{
		"claude-dm":     "DMs to Claude/OpenAI HTTP agent (no shell by default)",
		"copilot-shell": "DMs to Copilot CLI with shell action (trusted)",
		"local-llm":     "DMs to local HTTP LLM endpoint",
		"mock-echo":     "Offline mock transport + echo agent",
	}
}

// Get returns the raw YAML for a preset, or an error if unknown.
func Get(name string) ([]byte, error) {
	if data, ok := loadOverride(name); ok {
		return data, nil
	}
	switch name {
	case "claude-dm":
		return ClaudeDM, nil
	case "copilot-shell":
		return CopilotShell, nil
	case "local-llm":
		return LocalLLM, nil
	case "mock-echo":
		return MockEcho, nil
	default:
		return nil, fmt.Errorf("unknown preset %s", name)
	}
}

// loadOverride returns user/project preset overrides if present.
func loadOverride(name string) ([]byte, bool) {
	for _, path := range overridePaths(name) {
		if data, err := os.ReadFile(path); err == nil {
			return data, true
		}
	}
	return nil, false
}

func overridePaths(name string) []string {
	var paths []string
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".config", "buddy", "presets", name+".yaml"))
	}
	paths = append(paths, filepath.Join("presets", name+".yaml"))
	return paths
}
