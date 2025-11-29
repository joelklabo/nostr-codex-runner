package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"nostr-codex-runner/internal/codex"
	"nostr-codex-runner/internal/commands"
	"nostr-codex-runner/internal/config"
	"nostr-codex-runner/internal/nostrclient"
	"nostr-codex-runner/internal/store"

	"github.com/nbd-wtf/go-nostr/nip19"
)

var (
	processStart = time.Now()
	buildVer     = "dev"
	hostName     = "unknown"
	runnerPID    = os.Getpid()
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to config.yaml")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fatalf("load config: %v", err)
	}

	buildVer = buildVersion()
	if h, err := os.Hostname(); err == nil {
		hostName = h
	}

	level := slog.LevelInfo
	switch strings.ToLower(cfg.Logging.Level) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	var logFile *os.File
	writers := []io.Writer{os.Stdout}
	if cfg.Logging.File != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.Logging.File), 0o755); err != nil {
			fatalf("create log dir: %v", err)
		}
		logFile, err = os.OpenFile(cfg.Logging.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			fatalf("open log file: %v", err)
		}
		writers = append(writers, logFile)
		defer logFile.Close()
	}
	logger := slog.New(slog.NewTextHandler(io.MultiWriter(writers...), &slog.HandlerOptions{Level: level}))

	pubKey, err := cfg.GetRunnerPubKey()
	if err != nil {
		fatalf("derive pubkey: %v", err)
	}

	printBanner(cfg, pubKey, buildVer)

	st, err := store.New(cfg.Storage.Path)
	if err != nil {
		fatalf("open store: %v", err)
	}
	defer st.Close()

	runner := codex.New(cfg.Codex)
	client := nostrclient.New(cfg.Runner.PrivateKey, pubKey, cfg.Relays, cfg.Runner.AllowedPubkeys, st)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Info("nostr-codex-runner starting", slog.String("pubkey", pubKey), slog.Any("relays", cfg.Relays))

	errCh := make(chan error, 2)
	go func() {
		errCh <- client.Listen(ctx, func(msgCtx context.Context, msg nostrclient.IncomingMessage) {
			go handleMessage(msgCtx, logger, runner, client, st, cfg, msg, buildVer)
		})
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown requested")
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			fatalf("runtime error: %v", err)
		}
	}
}

func handleMessage(ctx context.Context, logger *slog.Logger, runner *codex.Runner, client *nostrclient.Client, st *store.Store, cfg *config.Config, msg nostrclient.IncomingMessage, version string) {
	cmd := commands.Parse(msg.Plaintext)
	sender := msg.SenderPubKey
	logger.Info("received DM", slog.String("from", sender), slog.String("cmd", cmd.Name))

	var newSession bool

	switch cmd.Name {
	case "help":
		_ = client.SendReply(ctx, sender, helpText())
		return
	case "status":
		stVal, ok, _ := st.Active(sender)
		if ok {
			_ = client.SendReply(ctx, sender, fmt.Sprintf("Active session: %s (updated %s)", stVal.SessionID, stVal.UpdatedAt.Format(time.RFC3339)))
		} else {
			_ = client.SendReply(ctx, sender, "No active session. Send a prompt to start one or /new to reset.")
		}
		return
	case "use":
		if cmd.Args == "" {
			_ = client.SendReply(ctx, sender, "Usage: /use <session-id>")
			return
		}
		if err := st.SaveActive(sender, cmd.Args); err != nil {
			_ = client.SendReply(ctx, sender, fmt.Sprintf("Failed to set active session: %v", err))
			return
		}
		_ = client.SendReply(ctx, sender, fmt.Sprintf("Switched to session %s", cmd.Args))
		return
	case "raw":
		if strings.TrimSpace(cmd.Args) == "" {
			_ = client.SendReply(ctx, sender, "Usage: /raw <shell command>")
			return
		}
		reply, exitCode := runRaw(ctx, cfg, cmd.Args)
		if exitCode == 0 {
			_ = client.SendReply(ctx, sender, reply)
			return
		}
		runErr := fmt.Errorf("/raw exit=%d: %s", exitCode, reply)
		handleRunError(ctx, logger, runner, client, cfg, sender, cmd.Args, runErr)
		return
	}

	// Determine session to use (if any) and prompt to run.
	var sessionID string
	var prompt string

	switch cmd.Name {
	case "new":
		_ = st.ClearActive(sender)
		_ = client.SendReply(ctx, sender, machineGreeting())
		if cmd.Args == "" {
			return
		}
		prompt = cmd.Args
		newSession = true
	case "run":
		prompt = cmd.Args
		if state, ok, _ := st.Active(sender); ok {
			if cfg.Runner.SessionTimeoutMins > 0 && time.Since(state.UpdatedAt) > time.Duration(cfg.Runner.SessionTimeoutMins)*time.Minute {
				logger.Info("active session expired", slog.String("from", sender), slog.String("session", state.SessionID))
				_ = st.ClearActive(sender)
			} else {
				sessionID = state.SessionID
			}
		}
	default:
		prompt = cmd.Args
	}

	if strings.TrimSpace(prompt) == "" {
		_ = client.SendReply(ctx, sender, "No prompt detected. Send text or /help for commands.")
		return
	}

	if sessionID == "" && strings.TrimSpace(cfg.Runner.InitialPrompt) != "" {
		prompt = cfg.Runner.InitialPrompt + "\n\n" + prompt
	}

	runCtx, cancel := runner.ContextWithTimeout(ctx)
	defer cancel()

	res, err := runner.Run(runCtx, sessionID, prompt)
	if err != nil {
		handleRunError(ctx, logger, runner, client, cfg, sender, prompt, err)
		return
	}

	// Persist active session for sender.
	if err := st.SaveActive(sender, res.SessionID); err != nil {
		logger.Error("failed to save session", slog.String("err", err.Error()))
	}

	meta := ""
	if newSession {
		meta = sessionMeta(cfg, version)
	}
	reply := formatReply(res, cfg.Runner.MaxReplyChars, meta)
	if cfg.Runner.AutoReply {
		if err := client.SendReply(ctx, sender, reply); err != nil {
			logger.Error("failed to send reply", slog.String("err", err.Error()))
		}
	}
}

func formatReply(res codex.Result, maxChars int, meta string) string {
	reply := res.Reply
	if maxChars > 0 {
		r := []rune(reply)
		if len(r) > maxChars {
			reply = string(r[:maxChars]) + "...\n(truncated)"
		}
	}
	if meta != "" {
		return fmt.Sprintf("session: %s\n%s\n\n%s", res.SessionID, meta, reply)
	}
	return fmt.Sprintf("session: %s\n\n%s", res.SessionID, reply)
}

func helpText() string {
	return "Commands:\n" +
		"/new [prompt]  - reset session; optionally run prompt in a brand new session\n" +
		"/use <session> - switch to an existing Codex session id\n" +
		"/raw <cmd>     - execute a shell command on the host\n" +
		"/status        - show active session\n" +
		"/help          - show this help\n" +
		"(any other text runs as a prompt in your active session)"
}

func printBanner(cfg *config.Config, pubKey string, version string) {
	if !isTTY() {
		return
	}

	cyan := "\033[36m"
	mag := "\033[35m"
	gray := "\033[90m"
	reset := "\033[0m"

	nsec := "(hidden)"
	if cfg.Runner.PrivateKey != "" {
		if enc, err := nostrEncodeNsec(cfg.Runner.PrivateKey); err == nil {
			nsec = enc
		}
	}
	logVal := "stdout"
	if cfg.Logging.File != "" {
		logVal = fmt.Sprintf("stdout + %s", cfg.Logging.File)
	}

	lines := []struct {
		label string
		value string
	}{
		{"pubkey", pubKey},
		{"nsec", nsec},
		{"relays", strings.Join(cfg.Relays, ", ")},
		{"cwd", cfg.Codex.WorkingDir},
		{"logs", logVal},
		{"version", version},
		{"pid", fmt.Sprintf("%d", runnerPID)},
	}

	const maxBannerWidth = 72

	title := "nostr-codex-runner"
	maxLen := len(title)
	for _, l := range lines {
		l.value = fitValue(l.label, l.value, maxBannerWidth-4)
		plain := fmt.Sprintf("%s  %s", l.label, l.value)
		if len(plain) > maxLen {
			maxLen = len(plain)
		}
	}
	padding := 2
	width := maxLen + padding*2
	if width > maxBannerWidth {
		width = maxBannerWidth
	}

	borderTop := fmt.Sprintf("%s╔%s╗%s", mag, strings.Repeat("═", width), reset)
	borderMid := fmt.Sprintf("%s╠%s╣%s", mag, strings.Repeat("═", width), reset)
	borderBot := fmt.Sprintf("%s╚%s╝%s", mag, strings.Repeat("═", width), reset)

	fmt.Println(borderTop)
	fmt.Printf("%s║%s%s%s║%s\n", mag, reset, center(title, width), mag, reset)
	fmt.Println(borderMid)
	for _, l := range lines {
		plain := fmt.Sprintf("%s  %s", l.label, l.value)
		pad := width - len(plain)
		if pad < 0 {
			pad = 0
		}
		visible := fmt.Sprintf("%s  %s%s%s", l.label, cyan, l.value, reset) + strings.Repeat(" ", pad)
		fmt.Printf("%s║%s%s%s%s║%s\n", mag, reset, gray, visible, mag, reset)
	}
	fmt.Println(borderBot)
	fmt.Printf("%sTip:%s DM /help to see commands.\n", gray, reset)
	fmt.Printf("%sTMUX:%s tmux attach -t nostr (logs to stdout)\n%s\n", gray, reset, reset)
}

func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func buildVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

func nostrEncodeNsec(sk string) (string, error) {
	enc, err := nip19.EncodePrivateKey(sk)
	if err != nil {
		return "", err
	}
	return enc, nil
}

func center(s string, width int) string {
	if len(s) >= width {
		return s
	}
	pad := width - len(s)
	left := pad / 2
	right := pad - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

func fitValue(label, value string, limit int) string {
	max := limit - len(label) - 2
	if max < 8 {
		max = 8
	}
	if len(value) > max {
		if max > 5 {
			value = value[:max-3] + "..."
		} else {
			value = value[:max]
		}
	}
	return value
}

func machineGreeting() string {
	host, _ := os.Hostname()
	cpu := runtime.NumCPU()
	cwd, _ := os.Getwd()
	free := humanDiskFree(cwd)
	return fmt.Sprintf("Starting fresh. Host=%s • CPUs=%d • Free@cwd=%s • cwd=%s", host, cpu, free, cwd)
}

func humanDiskFree(path string) string {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return "n/a"
	}
	free := st.Bavail * uint64(st.Bsize)
	return humanBytes(free)
}

func humanBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := unit, 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func runRaw(ctx context.Context, cfg *config.Config, command string) (string, int) {
	ctx, cancel := runnerTimeout(ctx, cfg)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-lc", command)
	if cfg.Codex.WorkingDir != "" {
		cmd.Dir = cfg.Codex.WorkingDir
	}
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	body := string(out)
	if body == "" {
		if err != nil {
			body = err.Error()
		} else {
			body = "(no output)"
		}
	}

	body = truncate(body, cfg.Runner.MaxReplyChars)
	return fmt.Sprintf("/raw exit=%d\n%s", exitCode, body), exitCode
}

func runnerTimeout(parent context.Context, cfg *config.Config) (context.Context, context.CancelFunc) {
	t := time.Duration(cfg.Codex.TimeoutSeconds) * time.Second
	if t == 0 {
		t = 15 * time.Minute
	}
	return context.WithTimeout(parent, t)
}

func truncate(s string, max int) string {
	if max <= 0 {
		return s
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "...\n(truncated)"
}

func handleRunError(ctx context.Context, logger *slog.Logger, runner *codex.Runner, client *nostrclient.Client, cfg *config.Config, sender string, prompt string, runErr error) {
	logger.Error("codex run failed", slog.String("from", sender), slog.String("err", runErr.Error()))

	logTail := tailLog(cfg.Logging.File, 8192)
	projects := projectList(cfg)
	diagPrompt := fmt.Sprintf(`You are HonkAI (ops/SRE). A nostr-codex-runner error occurred while handling a DM.

Sender: %s
Prompt that failed: %q
Error: %v
Log tail (latest): 
%s

Projects you can use: %s

Tasks:
1) Use bd to find or create an epic named "nostr-codex-runner errors" in the appropriate project (use dropdown/default project). If it exists, reuse it.
2) Create or update an issue for this specific error (keyed by the error text). Include reproduction hints, suspected root cause, and next steps.
3) Reply concisely (<=800 chars) summarizing the epic + issue status and what you’ll do next.

 Return ONLY the reply text to send back to the user.`, sender, prompt, runErr, logTail, projects)

	diagCtx, cancel := runner.ContextWithTimeout(ctx)
	defer cancel()

	res, err := runner.Run(diagCtx, "", diagPrompt)
	if err != nil {
		logger.Error("error handler failed", slog.String("from", sender), slog.String("err", err.Error()))
		_ = client.SendReply(ctx, sender, fmt.Sprintf("codex error (and recovery failed): %v", runErr))
		return
	}

	reply := formatReply(res, cfg.Runner.MaxReplyChars, "")
	if err := client.SendReply(ctx, sender, reply); err != nil {
		logger.Error("failed to send error reply", slog.String("err", err.Error()))
	}
}

func projectList(cfg *config.Config) string {
	if len(cfg.Projects) == 0 {
		return "(none configured)"
	}
	var b strings.Builder
	for i, p := range cfg.Projects {
		if i > 0 {
			b.WriteString("; ")
		}
		fmt.Fprintf(&b, "%s (%s)", p.ID, p.Path)
	}
	return b.String()
}

func tailLog(path string, maxBytes int64) string {
	if path == "" {
		return "(no log file configured)"
	}
	f, err := os.Open(path)
	if err != nil {
		return fmt.Sprintf("(log unreadable: %v)", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return fmt.Sprintf("(log stat error: %v)", err)
	}

	size := stat.Size()
	if size == 0 {
		return "(log empty)"
	}
	var start int64 = 0
	if size > maxBytes {
		start = size - maxBytes
	}
	if _, err := f.Seek(start, io.SeekStart); err != nil {
		return fmt.Sprintf("(log seek error: %v)", err)
	}
	buf, err := io.ReadAll(f)
	if err != nil {
		return fmt.Sprintf("(log read error: %v)", err)
	}
	return string(buf)
}

func sessionMeta(cfg *config.Config, version string) string {
	return fmt.Sprintf("runner pid=%d • host=%s • started=%s • cwd=%s • version=%s",
		runnerPID, hostName, processStart.Format(time.RFC3339), cfg.Codex.WorkingDir, version)
}

func fatalf(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
