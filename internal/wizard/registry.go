package wizard

// PromptKind enumerates the type of prompt.
type PromptKind string

const (
	PromptInput    PromptKind = "input"
	PromptPassword PromptKind = "password"
	PromptSelect   PromptKind = "select"
	PromptConfirm  PromptKind = "confirm"
)

// PromptSpec describes a question to ask.
type PromptSpec struct {
	Kind        PromptKind
	Label       string
	Default     string
	Options     []string
	Required    bool
	Description string
}

type TransportOption struct {
	Name        string
	Description string
	Prompts     []PromptSpec
}

type AgentOption struct {
	Name        string
	Description string
	Prompts     []PromptSpec
}

type ActionOption struct {
	Name          string
	Description   string
	DefaultEnable bool
	Prompts       []PromptSpec
}

type PresetOption struct {
	Name        string
	Description string
}

// Registry holds available options for the wizard.
type Registry struct {
	Transports []TransportOption
	Agents     []AgentOption
	Actions    []ActionOption
	Presets    []PresetOption
}

var defaultRegistry = Registry{
	Transports: []TransportOption{
		{Name: "nostr", Description: "Nostr DMs over relays"},
		{Name: "mock", Description: "Offline mock transport"},
	},
	Agents: []AgentOption{
		{Name: "http", Description: "Claude/OpenAI-style HTTP"},
		{Name: "copilotcli", Description: "GitHub Copilot CLI"},
		{Name: "echo", Description: "Echo responses (tests)"},
	},
	Actions: []ActionOption{
		{Name: "readfile", Description: "Read files from allowlisted roots", DefaultEnable: true},
		{Name: "shell", Description: "Execute shell commands (high risk)", DefaultEnable: false},
		{Name: "writefile", Description: "Write files to allowlisted roots", DefaultEnable: false},
	},
	Presets: []PresetOption{
		{Name: "claude-dm", Description: "Nostr DM to Claude/OpenAI agent"},
		{Name: "copilot-shell", Description: "Nostr DM to Copilot CLI + shell action"},
		{Name: "local-llm", Description: "Nostr DM to local HTTP LLM"},
		{Name: "mock-echo", Description: "Mock transport + echo agent (offline)"},
	},
}

// GetRegistry returns the default registry (copy).
func GetRegistry() Registry {
	return defaultRegistry
}

// SetRegistry overrides the global registry (primarily for tests/extensibility).
// Callers should restore the previous value after use to avoid leaking state across tests.
func SetRegistry(r Registry) {
	defaultRegistry = r
}
