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

func TestReadFileBinaryAndSizeLimits(t *testing.T) {
	dir := t.TempDir()
	binPath := filepath.Join(dir, "bin.dat")
	if err := os.WriteFile(binPath, []byte{0, 1, 2}, 0o644); err != nil {
		t.Fatalf("write bin: %v", err)
	}
	act := NewReadFile(Config{Roots: []string{dir}, MaxBytes: 2})
	args, _ := json.Marshal(map[string]string{"path": binPath})
	if _, err := act.Invoke(context.Background(), args); err == nil {
		t.Fatalf("expected binary rejection")
	}

	// Oversize text
	large := filepath.Join(dir, "large.txt")
	if err := os.WriteFile(large, []byte("abcdef"), 0o644); err != nil {
		t.Fatalf("write large: %v", err)
	}
	act2 := NewReadFile(Config{Roots: []string{dir}, MaxBytes: 3})
	args2, _ := json.Marshal(map[string]string{"path": large})
	if _, err := act2.Invoke(context.Background(), args2); err == nil {
		t.Fatalf("expected size error")
	}
}

func TestWriteFileSizeLimit(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f.txt")
	act := NewWriteFile(Config{Roots: []string{dir}, AllowWrite: true, MaxBytes: 3})
	args, _ := json.Marshal(map[string]string{"path": p, "content": "toolong"})
	if _, err := act.Invoke(context.Background(), args); err == nil {
		t.Fatalf("expected size error")
	}
}

func TestSafePathDefaults(t *testing.T) {
	if _, err := safePath(nil, ""); err == nil {
		t.Fatalf("expected error on empty path")
	}
	p, err := safePath(nil, "./relative.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(p) {
		t.Fatalf("expected absolute path, got %s", p)
	}
}

func TestIsBinaryHelper(t *testing.T) {
	if !isBinary([]byte{0x00, 0x01}) {
		t.Fatalf("expected binary detection")
	}
	if isBinary([]byte("hello")) {
		t.Fatalf("plain text should not be binary")
	}
}
