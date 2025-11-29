package main

import (
    "context"
    "errors"
    "flag"
    "fmt"
    "log/slog"
    "os"
    "os/signal"
    "strings"
    "syscall"
    "time"

    "nostr-codex-runner/internal/codex"
    "nostr-codex-runner/internal/commands"
    "nostr-codex-runner/internal/config"
    "nostr-codex-runner/internal/nostrclient"
    "nostr-codex-runner/internal/store"
)

func main() {
    configPath := flag.String("config", "config.yaml", "Path to config.yaml")
    flag.Parse()

    cfg, err := config.Load(*configPath)
    if err != nil {
        fatalf("load config: %v", err)
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
    logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

    pubKey, err := cfg.GetRunnerPubKey()
    if err != nil {
        fatalf("derive pubkey: %v", err)
    }

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

    errCh := make(chan error, 1)
    go func() {
        errCh <- client.Listen(ctx, func(msgCtx context.Context, msg nostrclient.IncomingMessage) {
            go handleMessage(msgCtx, logger, runner, client, st, cfg, msg)
        })
    }()

    select {
    case <-ctx.Done():
        logger.Info("shutdown requested")
    case err := <-errCh:
        if !errors.Is(err, context.Canceled) {
            fatalf("nostr listener error: %v", err)
        }
    }
}

func handleMessage(ctx context.Context, logger *slog.Logger, runner *codex.Runner, client *nostrclient.Client, st *store.Store, cfg *config.Config, msg nostrclient.IncomingMessage) {
    cmd := commands.Parse(msg.Plaintext)
    sender := msg.SenderPubKey
    logger.Info("received DM", slog.String("from", sender), slog.String("cmd", cmd.Name))

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
    }

    // Determine session to use (if any) and prompt to run.
    var sessionID string
    var prompt string

    switch cmd.Name {
    case "new":
        _ = st.ClearActive(sender)
        if cmd.Args == "" {
            _ = client.SendReply(ctx, sender, "Session reset. Send a prompt to start fresh.")
            return
        }
        prompt = cmd.Args
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

    runCtx, cancel := runner.ContextWithTimeout(ctx)
    defer cancel()

    res, err := runner.Run(runCtx, sessionID, prompt)
    if err != nil {
        logger.Error("codex run failed", slog.String("from", sender), slog.String("prompt", prompt), slog.String("err", err.Error()))
        _ = client.SendReply(ctx, sender, fmt.Sprintf("codex error: %v", err))
        return
    }

    // Persist active session for sender.
    if err := st.SaveActive(sender, res.SessionID); err != nil {
        logger.Error("failed to save session", slog.String("err", err.Error()))
    }

    reply := formatReply(res, cfg.Runner.MaxReplyChars)
    if cfg.Runner.AutoReply {
        if err := client.SendReply(ctx, sender, reply); err != nil {
            logger.Error("failed to send reply", slog.String("err", err.Error()))
        }
    }
}

func formatReply(res codex.Result, maxChars int) string {
    reply := res.Reply
    if maxChars > 0 {
        r := []rune(reply)
        if len(r) > maxChars {
            reply = string(r[:maxChars]) + "...\n(truncated)"
        }
    }
    return fmt.Sprintf("session: %s\n\n%s", res.SessionID, reply)
}

func helpText() string {
    return "Commands:\n" +
        "/new [prompt]  - reset session; optionally run prompt in a brand new session\n" +
        "/use <session> - switch to an existing Codex session id\n" +
        "/status        - show active session\n" +
        "/help          - show this help\n" +
        "(any other text runs as a prompt in your active session)"
}

func fatalf(msg string, args ...any) {
    fmt.Fprintf(os.Stderr, msg+"\n", args...)
    os.Exit(1)
}
