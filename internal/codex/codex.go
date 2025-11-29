package codex

import (
    "bufio"
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "os/exec"
    "strings"
    "time"

    "nostr-codex-runner/internal/config"
)

// Runner executes prompts through the codex CLI and extracts session metadata.
type Runner struct {
    cfg config.CodexConfig
}

// Result captures the most recent agent reply and session id.
type Result struct {
    SessionID string
    Reply     string
    RawLines  []string
}

// New creates a Runner with the provided config.
func New(cfg config.CodexConfig) *Runner {
    return &Runner{cfg: cfg}
}

// Run executes a prompt. If sessionID is empty, a new Codex session is started;
// otherwise the session is resumed.
func (r *Runner) Run(ctx context.Context, sessionID string, prompt string) (Result, error) {
    if strings.TrimSpace(prompt) == "" {
        return Result{}, errors.New("prompt cannot be empty")
    }

    args := make([]string, 0, 16)
    // Global flags
    if r.cfg.Approval != "" {
        args = append(args, "-a", r.cfg.Approval)
    }
    if r.cfg.Sandbox != "" {
        args = append(args, "--sandbox", r.cfg.Sandbox)
    }
    if r.cfg.Profile != "" {
        args = append(args, "--profile", r.cfg.Profile)
    }
    // Subcommand and options
    args = append(args, "exec", "--json")
    if r.cfg.SkipGitRepoCheck {
        args = append(args, "--skip-git-repo-check")
    }
    if len(r.cfg.ExtraArgs) > 0 {
        args = append(args, r.cfg.ExtraArgs...)
    }

    if sessionID == "" {
        // new session
    } else {
        args = append(args, "resume", sessionID)
    }

    args = append(args, prompt)

    cmd := exec.CommandContext(ctx, r.cfg.Binary, args...)
    if r.cfg.WorkingDir != "" {
        cmd.Dir = r.cfg.WorkingDir
    }

    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    if err := cmd.Run(); err != nil {
        return Result{}, fmt.Errorf("codex exec failed: %w; stderr: %s", err, stderr.String())
    }

    return parseCodexJSONL(stdout.Bytes())
}

// parseCodexJSONL extracts the session id and final agent message from Codex JSONL output.
func parseCodexJSONL(data []byte) (Result, error) {
    scanner := bufio.NewScanner(bytes.NewReader(data))
    scanner.Buffer(make([]byte, 0, 1024*64), 1024*1024) // allow larger lines

    var res Result
    for scanner.Scan() {
        line := scanner.Text()
        res.RawLines = append(res.RawLines, line)

        var evt struct {
            Type     string `json:"type"`
            ThreadID string `json:"thread_id"`
            Item     *struct {
                Type string `json:"type"`
                Text string `json:"text"`
            } `json:"item"`
            Error string `json:"error"`
        }

        if err := json.Unmarshal([]byte(line), &evt); err != nil {
            // Non-JSON lines are ignored but kept in raw log.
            continue
        }
        if evt.ThreadID != "" {
            res.SessionID = evt.ThreadID
        }
        if evt.Item != nil && evt.Item.Type == "agent_message" && evt.Item.Text != "" {
            res.Reply = evt.Item.Text
        }
        if evt.Error != "" {
            return res, fmt.Errorf("codex reported error: %s", evt.Error)
        }
    }
    if err := scanner.Err(); err != nil {
        return res, err
    }

    if res.SessionID == "" {
        return res, errors.New("could not find session id in codex output")
    }
    if res.Reply == "" {
        res.Reply = "(codex did not return a message)"
    }
    return res, nil
}

// ContextWithTimeout returns a context derived from parent with the configured timeout applied.
func (r *Runner) ContextWithTimeout(parent context.Context) (context.Context, context.CancelFunc) {
    t := time.Duration(r.cfg.TimeoutSeconds) * time.Second
    if t == 0 {
        t = 15 * time.Minute
    }
    return context.WithTimeout(parent, t)
}
