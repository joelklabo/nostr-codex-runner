package main

import (
	"os"
	"path/filepath"
	"testing"
	"strings"
)

func TestDefaultConfigPath_PrefersEnvNew(t *testing.T) {
	t.Setenv(envConfigNew, "/tmp/buddy.yaml")
	got := defaultConfigPath()
	if got != "/tmp/buddy.yaml" {
		t.Fatalf("expected env path, got %s", got)
	}
}

func TestDefaultConfigPath_CwdConfig(t *testing.T) {
	tmp := t.TempDir()
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "config.yaml"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(envConfigNew, "")

	got := defaultConfigPath()
	if got != "config.yaml" {
		t.Fatalf("expected cwd config.yaml, got %s", got)
	}
}

func TestDefaultConfigPath_HomeConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(envConfigNew, "")

	// no cwd config
	path := filepath.Join(home, ".config", "buddy", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	got := defaultConfigPath()
	if got != path {
		t.Fatalf("expected home buddy config, got %s", got)
	}
}

func TestLoadConfigWithPresets_Precedence(t *testing.T) {
	td := t.TempDir()
	home := filepath.Join(td, "home")
	cwd := filepath.Join(td, "cwd")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)

	// Home config
	homeCfg := filepath.Join(home, ".config", "buddy", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(homeCfg), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(homeCfg, minimalConfig("homeState"), 0o644); err != nil {
		t.Fatal(err)
	}

	// CWD config should win over home
	if err := os.WriteFile(filepath.Join(cwd, "config.yaml"), minimalConfig("cwdState"), 0o644); err != nil {
		t.Fatal(err)
	}

	// preset override should win if positional matches a preset name
	overrideDir := filepath.Join(home, ".config", "buddy", "presets")
	if err := os.MkdirAll(overrideDir, 0o755); err != nil {
		t.Fatal(err)
	}
	overridePath := filepath.Join(overrideDir, "mock-echo.yaml")
	overrideYAML := `
transports:
  - type: mock
    id: mock
agent:
  type: echo
actions:
  - type: shell
runner:
  private_key: "k"
  allowed_pubkeys: ["a"]
`
	if err := os.WriteFile(overridePath, []byte(overrideYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	origCwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(origCwd) })
	if err := os.Chdir(cwd); err != nil {
		t.Fatal(err)
	}

	cfg, preset, err := loadConfigWithPresets(defaultConfigPath(), "mock-echo")
	if err != nil {
		t.Fatalf("loadConfigWithPresets: %v", err)
	}
	if preset != "mock-echo" {
		t.Fatalf("expected preset name mock-echo, got %s", preset)
	}
	if cfg.Agent.Type != "echo" || len(cfg.Transports) == 0 || cfg.Transports[0].Type != "mock" {
		t.Fatalf("expected override preset loaded")
	}
}

func TestPresetsCommandPrintsOverrideYAML(t *testing.T) {
	td := t.TempDir()
	home := filepath.Join(td, "home")
	if err := os.MkdirAll(filepath.Join(home, ".config", "buddy", "presets"), 0o755); err != nil {
		t.Fatal(err)
	}
	overridePath := filepath.Join(home, ".config", "buddy", "presets", "mock-echo.yaml")
	overrideYAML := "transports: [{type: mock, id: mock}] agent: {type: echo}\nactions: [{type: readfile}]\nrunner: {private_key: \"k\", allowed_pubkeys: [\"a\"]}\n"
	if err := os.WriteFile(overridePath, []byte(overrideYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)

	out := captureStdout(func() {
		if err := runPresets([]string{"mock-echo", "--yaml"}); err != nil {
			t.Fatalf("runPresets: %v", err)
		}
	})
	if !strings.Contains(out, "allowed_pubkeys") {
		t.Fatalf("expected override yaml in output, got:\n%s", out)
	}
}

func minimalConfig(state string) []byte {
	return []byte(`runner:
  private_key: "k"
  allowed_pubkeys: ["a"]
storage:
  path: "` + state + `"
transports:
  - type: mock
    id: mock
agent:
  type: echo
actions:
  - type: shell
`)
}
