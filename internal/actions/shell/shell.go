// Package shell exposes a shell execution action with allowlist and truncation.
package shell

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Config controls the shell action.
type Config struct {
	Workdir        string
	Allowed        []string
	TimeoutSeconds int
	MaxOutput      int
}

// Action executes bash commands with allowlist + truncation.
type Action struct {
	cfg Config
}

func New(cfg Config) *Action {
	if cfg.TimeoutSeconds == 0 {
		cfg.TimeoutSeconds = 30
	}
	if cfg.MaxOutput == 0 {
		cfg.MaxOutput = 8000
	}
	return &Action{cfg: cfg}
}

func (a *Action) Name() string { return "shell" }

func (a *Action) Capabilities() []string { return []string{"shell:exec"} }

func (a *Action) Invoke(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var payload struct {
		Command string `json:"command"`
	}
	if len(args) == 0 {
		return nil, errors.New("missing command")
	}
	if err := json.Unmarshal(args, &payload); err != nil {
		// allow raw string
		var s string
		if err2 := json.Unmarshal(args, &s); err2 != nil {
			return nil, fmt.Errorf("decode args: %w", err)
		}
		payload.Command = s
	}
	cmdStr := strings.TrimSpace(payload.Command)
	if cmdStr == "" {
		return nil, errors.New("empty command")
	}
	if !a.allowed(cmdStr) {
		return nil, fmt.Errorf("command not allowed")
	}

	timeout := time.Duration(a.cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(cctx, "bash", "-lc", cmdStr)
	if a.cfg.Workdir != "" {
		cmd.Dir = a.cfg.Workdir
	}
	out, err := cmd.CombinedOutput()
	if cctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("command timeout")
	}
	text := string(out)
	text = truncate(text, a.cfg.MaxOutput)
	encoded, _ := json.Marshal(text)
	if err != nil {
		return json.RawMessage(fmt.Sprintf("\"/shell exit=%d\\n%s\"", exitCode(err), escapeJSONString(text))), nil
	}
	return json.RawMessage(encoded), nil
}

func (a *Action) allowed(cmd string) bool {
	if len(a.cfg.Allowed) == 0 {
		return true
	}
	for _, prefix := range a.cfg.Allowed {
		if strings.HasPrefix(cmd, prefix) {
			return true
		}
	}
	return false
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	if max > 3 {
		return s[:max-3] + "..."
	}
	return s[:max]
}

func exitCode(err error) int {
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return 1
}

// escapeJSONString is a minimal string escaper for JSON string output.
func escapeJSONString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
