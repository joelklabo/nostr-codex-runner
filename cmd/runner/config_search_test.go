package main

import (
	"os"
	"path/filepath"
	"testing"
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
