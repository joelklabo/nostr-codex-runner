package fs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Config for filesystem actions.
type Config struct {
	Roots      []string
	MaxBytes   int64
	AllowWrite bool
}

// ReadFile action reads text files under allowed roots.
type ReadFile struct {
	cfg Config
}

func NewReadFile(cfg Config) *ReadFile {
	if cfg.MaxBytes == 0 {
		cfg.MaxBytes = 64 * 1024
	}
	return &ReadFile{cfg: cfg}
}

func (r *ReadFile) Name() string           { return "readfile" }
func (r *ReadFile) Capabilities() []string { return []string{"fs:read"} }

func (r *ReadFile) Invoke(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var payload struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(args, &payload); err != nil {
		return nil, fmt.Errorf("decode args: %w", err)
	}
	p, err := r.safePath(payload.Path)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	limited := io.LimitReader(f, r.cfg.MaxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > r.cfg.MaxBytes {
		return nil, fmt.Errorf("file too large (>%d bytes)", r.cfg.MaxBytes)
	}
	if isBinary(data) {
		return nil, errors.New("refusing to return binary content")
	}
	encoded, _ := json.Marshal(string(data))
	return encoded, nil
}

// WriteFile writes text content to a file under allowed roots.
type WriteFile struct {
	cfg Config
}

func NewWriteFile(cfg Config) *WriteFile {
	if cfg.MaxBytes == 0 {
		cfg.MaxBytes = 64 * 1024
	}
	return &WriteFile{cfg: cfg}
}

func (w *WriteFile) Name() string           { return "writefile" }
func (w *WriteFile) Capabilities() []string { return []string{"fs:write"} }

func (w *WriteFile) Invoke(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	if !w.cfg.AllowWrite {
		return nil, errors.New("write not permitted")
	}
	var payload struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(args, &payload); err != nil {
		return nil, fmt.Errorf("decode args: %w", err)
	}
	p, err := w.safePath(payload.Path)
	if err != nil {
		return nil, err
	}
	if int64(len(payload.Content)) > w.cfg.MaxBytes {
		return nil, fmt.Errorf("content too large")
	}
	if err := os.WriteFile(p, []byte(payload.Content), 0o644); err != nil {
		return nil, err
	}
	return json.RawMessage(`"ok"`), nil
}

// safePath ensures path is within allowed roots.
func (r *ReadFile) safePath(p string) (string, error)  { return safePath(r.cfg.Roots, p) }
func (w *WriteFile) safePath(p string) (string, error) { return safePath(w.cfg.Roots, p) }

func safePath(roots []string, p string) (string, error) {
	if p == "" {
		return "", errors.New("path required")
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	for _, root := range roots {
		if root == "" {
			continue
		}
		rAbs, err := filepath.Abs(root)
		if err != nil {
			continue
		}
		if strings.HasPrefix(abs, rAbs) {
			return abs, nil
		}
	}
	if len(roots) == 0 {
		return abs, nil
	}
	return "", fmt.Errorf("path outside allowed roots")
}

func isBinary(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return false
}
