package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/nbd-wtf/go-nostr"
	"gopkg.in/yaml.v3"
)

// Config holds the runtime configuration loaded from config.yaml.
type Config struct {
	Relays  []string      `yaml:"relays"`
	Runner  RunnerConfig  `yaml:"runner"`
	Codex   CodexConfig   `yaml:"codex"`
	Storage StorageConfig `yaml:"storage"`
	Logging LoggingConfig `yaml:"logging"`
}

// RunnerConfig controls Nostr-facing behaviour.
type RunnerConfig struct {
	PrivateKey         string   `yaml:"private_key"`
	AllowedPubkeys     []string `yaml:"allowed_pubkeys"`
	AutoReply          bool     `yaml:"auto_reply"`
	MaxReplyChars      int      `yaml:"max_reply_chars"`
	SessionTimeoutMins int      `yaml:"session_timeout_minutes"`
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
	Level string `yaml:"level"`
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

	cfg.applyDefaults()
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
	if len(c.Relays) == 0 {
		return errors.New("at least one relay is required")
	}
	if c.Runner.PrivateKey == "" {
		return errors.New("runner.private_key is required")
	}
	if len(c.Runner.AllowedPubkeys) == 0 {
		return errors.New("runner.allowed_pubkeys must contain at least one key")
	}
	if c.Codex.Binary == "" {
		return errors.New("codex.binary is required (e.g., 'codex')")
	}
	if c.Storage.Path == "" {
		return errors.New("storage.path is required")
	}
	return nil
}

func (c *Config) applyDefaults() {
	if c.Runner.MaxReplyChars == 0 {
		c.Runner.MaxReplyChars = 8000
	}
	if c.Runner.SessionTimeoutMins == 0 {
		c.Runner.SessionTimeoutMins = 240
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
	// Ensure keys are lowercase to avoid mismatches.
	for i, pk := range c.Runner.AllowedPubkeys {
		c.Runner.AllowedPubkeys[i] = strings.ToLower(pk)
	}
}
