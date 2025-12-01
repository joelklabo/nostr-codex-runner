package presets

import (
	_ "embed"

	"github.com/joelklabo/buddy/internal/config"
)

//go:embed data/claude-dm.yaml
var ClaudeDM []byte

//go:embed data/copilot-shell.yaml
var CopilotShell []byte

//go:embed data/local-llm.yaml
var LocalLLM []byte

//go:embed data/mock-echo.yaml
var MockEcho []byte

// PresetDeps returns declared prerequisites for built-in presets.
func PresetDeps() map[string][]config.Dep {
	return map[string][]config.Dep{
		"copilot-shell": {
			{Name: "copilot", Type: "binary", Hint: "Install GitHub Copilot CLI: https://github.com/github/copilot-cli"},
			{Name: "https://api.github.com", Type: "url", Optional: true, Hint: "Copilot auth check"},
			{Name: ".", Type: "dirwrite", Optional: true, Hint: "Workspace must be writable"},
		},
		"claude-dm": {
			{Name: "curl", Type: "binary", Optional: true, Hint: "Used for simple HTTP checks"},
			{Name: "https://api.anthropic.com", Type: "url", Optional: true, Hint: "Claude endpoint reachability"},
			{Name: ".", Type: "dirwrite", Optional: true, Hint: "Workspace must be writable"},
		},
		"local-llm": {
			{Name: "curl", Type: "binary", Optional: true, Hint: "Useful for hitting local endpoints"},
			{Name: "127.0.0.1:11434", Type: "port", Optional: true, Hint: "Example local LLM port (ollama)"},
			{Name: ".", Type: "dirwrite", Optional: true, Hint: "Workspace must be writable"},
		},
	}
}
