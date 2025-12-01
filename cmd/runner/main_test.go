package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/joelklabo/buddy/internal/config"
)

func TestParseSubcommand(t *testing.T) {
	cmd, rest := parseSubcommand([]string{"version"})
	if cmd != "version" || len(rest) != 0 {
		t.Fatalf("parse subcommand failed")
	}
	cmd, rest = parseSubcommand([]string{"presets"})
	if cmd != "presets" || len(rest) != 0 {
		t.Fatalf("expected presets routing")
	}
	cmd, rest = parseSubcommand([]string{"-config", "x"})
	if cmd != "run" || len(rest) != 2 {
		t.Fatalf("expected run fallback")
	}
}

func TestDefaultConfigPathEnv(t *testing.T) {
	if err := os.Setenv(envConfigNew, "/tmp/cfg"); err != nil {
		t.Fatalf("set env: %v", err)
	}
	defer func() { _ = os.Unsetenv(envConfigNew) }()
	if got := defaultConfigPath(); got != "/tmp/cfg" {
		t.Fatalf("expected env path, got %s", got)
	}
}

func TestUsageDoesNotPanic(t *testing.T) {
	usage()
}

func TestDefaultConfigPathFallback(t *testing.T) {
	t.Setenv(envConfigNew, "")

	got := defaultConfigPath()
	if got == "" {
		t.Fatalf("unexpected empty path")
	}
}

func TestParseSubcommandDefault(t *testing.T) {
	cmd, rest := parseSubcommand([]string{})
	if cmd != "run" || len(rest) != 0 {
		t.Fatalf("expected run default, got %s", cmd)
	}
}

func TestSetupLoggerWritesFile(t *testing.T) {
	td := t.TempDir()
	cfg := &config.Config{
		Logging: config.LoggingConfig{
			Level:  "debug",
			File:   filepath.Join(td, "runner.log"),
			Format: "json",
		},
	}
	logger := setupLogger(cfg)
	logger.Info("hello", "k", "v")

	data, err := os.ReadFile(cfg.Logging.File)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if !strings.Contains(string(data), `"msg":"hello"`) {
		t.Fatalf("log content unexpected: %s", string(data))
	}
}

func TestPrintBannerOutputs(t *testing.T) {
	cfg := &config.Config{}
	out := captureStdout(func() { printBanner(cfg, "pub", "v1.2.3") })
	if !strings.Contains(out, "buddy v1.2.3") {
		t.Fatalf("banner missing version: %s", out)
	}
}

func TestBuildVersionNonEmpty(t *testing.T) {
	if buildVersion() == "" {
		t.Fatalf("buildVersion should not be empty")
	}
}

func TestRunContextStartsAndCancels(t *testing.T) {
	td := t.TempDir()
	cfgPath := filepath.Join(td, "config.yaml")
	cfgYAML := `
runner:
  private_key: "abcd"
  allowed_pubkeys: ["1234"]
storage:
  path: "` + filepath.Join(td, "state.db") + `"
transports:
  - type: mock
    id: mock
agent:
  type: echo
actions:
  - type: shell
projects:
  - id: default
    name: default
    path: .
`
	if err := os.WriteFile(cfgPath, []byte(cfgYAML), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()
	if err := runContext(ctx, []string{"-config", cfgPath}); err != nil && err != context.Canceled {
		t.Fatalf("runContext err: %v", err)
	}
}

func TestRunContextWithMockPreset(t *testing.T) {
	td := t.TempDir()
	t.Setenv("HOME", td)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(150 * time.Millisecond)
		cancel()
	}()
	if err := runContext(ctx, []string{"-skip-check", "mock-echo"}); err != nil && err != context.Canceled {
		t.Fatalf("runContext preset mock-echo err: %v", err)
	}
	if _, err := os.Stat(filepath.Join(td, ".buddy", "state.db")); err != nil {
		t.Fatalf("expected state db under HOME: %v", err)
	}
}

func TestRunCheckUsesConfigDeps(t *testing.T) {
	td := t.TempDir()
	cfgPath := filepath.Join(td, "config.yaml")
	statePath := filepath.Join(td, "state.db")
	requiredEnv := "BUDDY_CHECK_TEST_ENV"
	t.Setenv(requiredEnv, "ok")

	cfgYAML := `
runner:
  private_key: ""
  allowed_pubkeys: []
storage:
  path: "` + statePath + `"
transports:
  - type: mock
    id: mock
agent:
  type: echo
actions: []
deps:
  transports:
    mock:
      - name: "` + requiredEnv + `"
        type: env
        optional: false
        hint: "setenv in test"
      - name: "` + td + `"
        type: dirwrite
        optional: false
        hint: "temp dir writable"
`
	if err := os.WriteFile(cfgPath, []byte(cfgYAML), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if err := runCheck([]string{"-config", cfgPath}); err != nil {
		t.Fatalf("runCheck err: %v", err)
	}
}

func TestRunCheckFailsOnMissingRequired(t *testing.T) {
	td := t.TempDir()
	cfgPath := filepath.Join(td, "config.yaml")
	statePath := filepath.Join(td, "state.db")

	cfgYAML := `
runner:
  private_key: ""
  allowed_pubkeys: []
storage:
  path: "` + statePath + `"
transports:
  - type: mock
    id: mock
agent:
  type: echo
actions: []
deps:
  transports:
    mock:
      - name: "/definitely/missing/path/for/test"
        type: file
        optional: false
        hint: "intentionally missing"
`
	if err := os.WriteFile(cfgPath, []byte(cfgYAML), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if err := runCheck([]string{"-config", cfgPath}); err == nil {
		t.Fatalf("expected runCheck to fail on missing required dep")
	}
}

func TestRunCheckJSONOutput(t *testing.T) {
	td := t.TempDir()
	cfgPath := filepath.Join(td, "config.yaml")
	statePath := filepath.Join(td, "state.db")
	envName := "BUDDY_CHECK_JSON_ENV"
	t.Setenv(envName, "ok")

	cfgYAML := `
runner:
  private_key: ""
  allowed_pubkeys: []
storage:
  path: "` + statePath + `"
transports:
  - type: mock
    id: mock
agent:
  type: echo
actions: []
deps:
  transports:
    mock:
      - name: "` + envName + `"
        type: env
        optional: false
        hint: "set by test"
`
	if err := os.WriteFile(cfgPath, []byte(cfgYAML), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	out := captureStdout(func() {
		if err := runCheck([]string{"-config", cfgPath, "-json"}); err != nil {
			t.Fatalf("runCheck err: %v", err)
		}
	})

	var results []map[string]any
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("unmarshal json: %v\noutput: %s", err, out)
	}
	if len(results) == 0 {
		t.Fatalf("expected at least one result")
	}
	if results[0]["status"] == "" {
		t.Fatalf("expected status field in result: %#v", results[0])
	}
}

func TestRunContextUsesBuddyConfigEnv(t *testing.T) {
	td := t.TempDir()
	envCfg := filepath.Join(td, "env-config.yaml")
	envDB := filepath.Join(td, "env.db")
	cwd := filepath.Join(td, "cwd")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	// Env config (should be preferred)
	envYAML := `
runner:
  private_key: "envk"
  allowed_pubkeys: ["env"]
storage:
  path: "` + envDB + `"
transports:
  - type: mock
    id: mock
agent:
  type: echo
actions:
  - type: shell
projects:
  - id: default
    name: default
    path: .
`
	if err := os.WriteFile(envCfg, []byte(envYAML), 0o644); err != nil {
		t.Fatalf("write env config: %v", err)
	}
	// CWD config (should be ignored because env wins)
	cwdCfg := filepath.Join(cwd, "config.yaml")
	cwdDB := filepath.Join(cwd, "cwd.db")
	cwdYAML := strings.ReplaceAll(envYAML, envDB, cwdDB)
	if err := os.WriteFile(cwdCfg, []byte(cwdYAML), 0o644); err != nil {
		t.Fatalf("write cwd config: %v", err)
	}

	t.Setenv(envConfigNew, envCfg)
	origCwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(origCwd) })
	if err := os.Chdir(cwd); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(150 * time.Millisecond)
		cancel()
	}()
	if err := runContext(ctx, []string{"-skip-check"}); err != nil && err != context.Canceled {
		t.Fatalf("runContext err: %v", err)
	}
	if _, err := os.Stat(envDB); err != nil {
		t.Fatalf("expected env db created, got %v", err)
	}
	if _, err := os.Stat(cwdDB); err == nil {
		t.Fatalf("cwd db should not have been used")
	}
}

func TestRunContextWithHealthAndMetrics(t *testing.T) {
	td := t.TempDir()
	cfgPath := filepath.Join(td, "config.yaml")
	cfgYAML := `
runner:
  private_key: "abcd"
  allowed_pubkeys: ["1234"]
storage:
  path: "` + filepath.Join(td, "state.db") + `"
transports:
  - type: mock
    id: mock
agent:
  type: echo
actions:
  - type: shell
projects:
  - id: default
    name: default
    path: .
`
	if err := os.WriteFile(cfgPath, []byte(cfgYAML), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(150 * time.Millisecond)
		cancel()
	}()
	if err := runContext(ctx, []string{"-config", cfgPath, "-health-listen", "127.0.0.1:0", "-metrics-listen", "127.0.0.1:0", "-skip-check"}); err != nil && err != context.Canceled {
		t.Fatalf("runContext err: %v", err)
	}
}

func captureStdout(fn func()) string {
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = orig
	}()
	fn()
	_ = w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}
