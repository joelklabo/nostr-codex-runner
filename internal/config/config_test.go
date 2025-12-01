package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}

func TestLoadAppliesDefaults(t *testing.T) {
	cfgPath := writeTempConfig(t, `
runner:
  private_key: "abcd"
  allowed_pubkeys: ["1234"]
storage:
  path: "./state.db"
transports:
  - type: mock
    id: mock
agent:
  type: echo
actions:
  - type: shell
projects:
  - id: default
    path: .
    name: default
`)
	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Runner.ProfileName == "" || cfg.Runner.ProfileImage == "" {
		t.Fatalf("expected profile defaults applied")
	}
	if cfg.Runner.InitialPrompt == "" {
		t.Fatalf("initial prompt default missing")
	}
	if cfg.Agent.Type != "echo" {
		t.Fatalf("agent type wrong: %s", cfg.Agent.Type)
	}
	if len(cfg.Actions) != 1 || cfg.Actions[0].Type != "shell" {
		t.Fatalf("actions not parsed")
	}
}

func TestLoadAddsReadfileWhenActionsEmpty(t *testing.T) {
	cfgPath := writeTempConfig(t, `
runner:
  private_key: "abcd"
  allowed_pubkeys: ["1234"]
storage:
  path: "./state.db"
transports:
  - type: mock
    id: mock
agent:
  type: echo
actions: []
projects:
  - id: default
    path: .
    name: default
`)
	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(cfg.Actions) == 0 || cfg.Actions[0].Type != "readfile" {
		t.Fatalf("expected default readfile action, got %#v", cfg.Actions)
	}
}

func TestLoadMissingRequiredFails(t *testing.T) {
	cfgPath := writeTempConfig(t, `
runner:
  private_key: ""
  allowed_pubkeys: []
storage:
  path: ""
transports: []
agent:
  type: ""
actions: []
projects: []
`)
	if _, err := Load(cfgPath); err == nil {
		t.Fatalf("expected validation error")
	}
}
