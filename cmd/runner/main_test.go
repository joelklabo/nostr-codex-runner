package main

import (
	"os"
	"testing"
)

func TestParseSubcommand(t *testing.T) {
	cmd, rest := parseSubcommand([]string{"version"})
	if cmd != "version" || len(rest) != 0 {
		t.Fatalf("parse subcommand failed")
	}
	cmd, rest = parseSubcommand([]string{"-config", "x"})
	if cmd != "run" || len(rest) != 2 {
		t.Fatalf("expected run fallback")
	}
}

func TestDefaultConfigPathEnv(t *testing.T) {
	os.Setenv(envConfig, "/tmp/cfg")
	defer os.Unsetenv(envConfig)
	if got := defaultConfigPath(); got != "/tmp/cfg" {
		t.Fatalf("expected env path, got %s", got)
	}
}
