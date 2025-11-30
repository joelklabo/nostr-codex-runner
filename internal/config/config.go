package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"gopkg.in/yaml.v3"
)

// Config holds the runtime configuration loaded from config.yaml.
type Config struct {
	Relays   []string      `yaml:"relays"`
	Runner   RunnerConfig  `yaml:"runner"`
	Codex    CodexConfig   `yaml:"codex"`
	Storage  StorageConfig `yaml:"storage"`
	Logging  LoggingConfig `yaml:"logging"`
	Projects []Project     `yaml:"projects"`

	Transports []TransportConfig `yaml:"transports"`
	Agent      AgentConfig       `yaml:"agent"`
	Actions    []ActionConfig    `yaml:"actions"`
}

// RunnerConfig controls Nostr-facing behaviour.
type RunnerConfig struct {
	PrivateKey         string   `yaml:"private_key"`
	AllowedPubkeys     []string `yaml:"allowed_pubkeys"`
	AutoReply          bool     `yaml:"auto_reply"`
	MaxReplyChars      int      `yaml:"max_reply_chars"`
	SessionTimeoutMins int      `yaml:"session_timeout_minutes"`
	InitialPrompt      string   `yaml:"initial_prompt"`
	ProfileName        string   `yaml:"profile_name"`
	ProfileImage       string   `yaml:"profile_image"`
}

// CodexConfig controls how we invoke the codex CLI.
type CodexConfig struct {
	Binary           string   `yaml:"binary"`
	Sandbox          string   `yaml:"sandbox"`
	Approval         string   `yaml:"approval"`
	Profile          string   `yaml:"profile"`
	WorkingDir       string   `yaml:"working_dir"`
	ExtraArgs        []string `yaml:"extra_args"`
	SkipGitRepoCheck bool     `yaml:"skip_git_repo_check"`
	TimeoutSeconds   int      `yaml:"timeout_seconds"`
}

// StorageConfig controls persistence.
type StorageConfig struct {
	Path string `yaml:"path"`
}

// LoggingConfig controls log level.
type LoggingConfig struct {
	Level  string `yaml:"level"`
	File   string `yaml:"file"`
	Format string `yaml:"format"`
}

// Project represents a bd workspace (path containing a .beads dir).
type Project struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

// TransportConfig is a generic transport entry.
type TransportConfig struct {
	Type string `yaml:"type"`
	ID   string `yaml:"id"`

	// Nostr-specific fields (used when type=nostr)
	Relays         []string `yaml:"relays"`
	PrivateKey     string   `yaml:"private_key"`
	AllowedPubkeys []string `yaml:"allowed_pubkeys"`
}

// AgentConfig holds agent selection and backend config.
type AgentConfig struct {
	Type string `yaml:"type"`
	// Codex CLI fields (type=codexcli)
	Codex CodexConfig `yaml:"codex"`
}

// ActionConfig defines an action plugin instance.
type ActionConfig struct {
	Type         string   `yaml:"type"`
	Name         string   `yaml:"name"`
	Workdir      string   `yaml:"workdir"`
	Allowed      []string `yaml:"allowed"`
	TimeoutSecs  int      `yaml:"timeout_seconds"`
	MaxOutput    int      `yaml:"max_output"`
	Roots        []string `yaml:"roots"`
	AllowWrite   bool     `yaml:"allow_write"`
	MaxBytes     int64    `yaml:"max_bytes"`
	Capabilities []string `yaml:"capabilities"`
	Description  string   `yaml:"description"`
}

// Load reads and validates configuration from the provided path.
func Load(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	baseDir := filepath.Dir(path)
	cfg.applyDefaults(baseDir)
	cfg.Storage.Path = expandPath(cfg.Storage.Path)
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// GetRunnerPubKey derives the runner's public key from its private key.
func (c *Config) GetRunnerPubKey() (string, error) {
	if c.Runner.PrivateKey == "" {
		return "", errors.New("runner.private_key is required")
	}
	pub, err := nostr.GetPublicKey(c.Runner.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("derive pubkey: %w", err)
	}
	return pub, nil
}

// Validate ensures the config is usable.
func (c *Config) Validate() error {
	if c.Runner.PrivateKey == "" {
		return errors.New("runner.private_key is required")
	}
	if len(c.Runner.AllowedPubkeys) == 0 {
		return errors.New("runner.allowed_pubkeys must contain at least one key")
	}
	if c.Storage.Path == "" {
		return errors.New("storage.path is required")
	}
	if len(c.Transports) == 0 {
		return errors.New("at least one transport is required")
	}
	if c.Agent.Type == "" {
		return errors.New("agent.type is required")
	}
	if len(c.Actions) == 0 {
		return errors.New("at least one action is required")
	}
	if len(c.Projects) == 0 {
		return errors.New("at least one project must be configured")
	}
	for _, p := range c.Projects {
		if p.Path == "" {
			return fmt.Errorf("project %s has empty path", p.ID)
		}
	}
	return nil
}

func (c *Config) applyDefaults(baseDir string) {
	if c.Runner.MaxReplyChars == 0 {
		c.Runner.MaxReplyChars = 8000
	}
	if c.Runner.SessionTimeoutMins == 0 {
		c.Runner.SessionTimeoutMins = 240
	}
	if strings.TrimSpace(c.Runner.InitialPrompt) == "" {
		c.Runner.InitialPrompt = "You are an AI agent with shell access to this machine via Codex. Be concise, be careful, and always explain what you plan to do before running commands. Ask for confirmation before risky actions."
	}
	if c.Runner.ProfileName == "" {
		c.Runner.ProfileName = "nostr-codex-runner"
	}
	if c.Runner.ProfileImage == "" {
		c.Runner.ProfileImage = "https://raw.githubusercontent.com/joelklabo/nostr-codex-runner/main/assets/social-preview.svg"
	}
	if c.Codex.Sandbox == "" {
		c.Codex.Sandbox = "danger-full-access"
	}
	if c.Codex.Approval == "" {
		c.Codex.Approval = "never"
	}
	if c.Codex.Binary == "" {
		c.Codex.Binary = "codex"
	}
	if c.Codex.WorkingDir == "" {
		if home, err := os.UserHomeDir(); err == nil {
			c.Codex.WorkingDir = home
		} else {
			c.Codex.WorkingDir = "."
		}
	}
	if c.Codex.TimeoutSeconds == 0 {
		c.Codex.TimeoutSeconds = 900 // 15 minutes
	}
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "text"
	}
	if c.Logging.File == "" {
		if home, err := os.UserHomeDir(); err == nil {
			c.Logging.File = filepath.Join(home, ".nostr-codex", "runner.log")
		}
	}

	if len(c.Projects) == 0 {
		wd := baseDir
		if abs, err := filepath.Abs(wd); err == nil {
			wd = abs
		}
		c.Projects = []Project{{ID: "default", Name: "Default", Path: wd}}
	}
	for i := range c.Projects {
		if c.Projects[i].ID == "" {
			c.Projects[i].ID = fmt.Sprintf("project-%d", i+1)
		}
		if c.Projects[i].Name == "" {
			c.Projects[i].Name = c.Projects[i].ID
		}
		if c.Projects[i].Path == "" {
			c.Projects[i].Path = baseDir
		}
		if abs, err := filepath.Abs(c.Projects[i].Path); err == nil {
			c.Projects[i].Path = abs
		}
	}
	if c.Storage.Path == "" {
		if home, err := os.UserHomeDir(); err == nil {
			c.Storage.Path = filepath.Join(home, ".nostr-codex", "state.db")
		}
	}
	if c.Logging.File != "" {
		c.Logging.File = expandPath(c.Logging.File)
	}
	// Normalize allowed pubkeys to lowercase hex.
	for i, pk := range c.Runner.AllowedPubkeys {
		c.Runner.AllowedPubkeys[i] = normalizePubkey(pk)
	}

	// Defaults for plugin schema (backward compat)
	if len(c.Transports) == 0 {
		c.Transports = []TransportConfig{{
			Type:           "nostr",
			ID:             "nostr",
			Relays:         c.Relays,
			PrivateKey:     c.Runner.PrivateKey,
			AllowedPubkeys: c.Runner.AllowedPubkeys,
		}}
	}
	if c.Agent.Type == "" {
		c.Agent.Type = "codexcli"
		c.Agent.Codex = c.Codex
	}
	if len(c.Actions) == 0 {
		c.Actions = []ActionConfig{{
			Type:        "shell",
			Name:        "shell",
			Workdir:     c.Codex.WorkingDir,
			TimeoutSecs: c.Codex.TimeoutSeconds,
			MaxOutput:   c.Runner.MaxReplyChars,
			Allowed:     []string{},
		}}
	}
}

func expandPath(p string) string {
	if p == "" {
		return p
	}
	if strings.HasPrefix(p, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			p = filepath.Join(home, strings.TrimPrefix(p, "~"))
		}
	}
	return filepath.Clean(os.ExpandEnv(p))
}

func normalizePubkey(pk string) string {
	pk = strings.TrimSpace(pk)
	pk = strings.ToLower(pk)
	if strings.HasPrefix(pk, "npub") {
		kind, data, err := nip19.Decode(pk)
		if err == nil && kind == "npub" {
			if hex, ok := data.(string); ok {
				return strings.ToLower(hex)
			}
		}
	}
	return pk
}
