package fs

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestReadFileAllowed(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(p, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write seed file: %v", err)
	}

	act := NewReadFile(Config{Roots: []string{dir}})
	args, _ := json.Marshal(map[string]string{"path": p})
	out, err := act.Invoke(context.Background(), args)
	if err != nil {
		t.Fatalf("invoke err: %v", err)
	}
	if string(out) != "\"hello\"" {
		t.Fatalf("unexpected out: %s", out)
	}
}

func TestReadFileDenied(t *testing.T) {
	dir := t.TempDir()
	act := NewReadFile(Config{Roots: []string{dir}})
	args, _ := json.Marshal(map[string]string{"path": "/etc/passwd"})
	if _, err := act.Invoke(context.Background(), args); err == nil {
		t.Fatalf("expected deny")
	}
}

func TestWriteFileAllowed(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f.txt")
	act := NewWriteFile(Config{Roots: []string{dir}, AllowWrite: true})
	args, _ := json.Marshal(map[string]string{"path": p, "content": "ok"})
	if _, err := act.Invoke(context.Background(), args); err != nil {
		t.Fatalf("write err: %v", err)
	}
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(data) != "ok" {
		t.Fatalf("unexpected data: %s", data)
	}
}

func TestWriteFileDenied(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f.txt")
	act := NewWriteFile(Config{Roots: []string{dir}, AllowWrite: false})
	args, _ := json.Marshal(map[string]string{"path": p, "content": "ok"})
	if _, err := act.Invoke(context.Background(), args); err == nil {
		t.Fatalf("expected write deny")
	}
}
