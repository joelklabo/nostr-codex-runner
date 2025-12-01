package main

import (
	"bytes"
	"context"
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
