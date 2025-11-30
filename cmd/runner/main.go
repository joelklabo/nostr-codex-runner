package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"nostr-codex-runner/internal/app"
	"nostr-codex-runner/internal/config"
	"nostr-codex-runner/internal/store"
)

var (
	processStart = time.Now()
	buildVer     = "dev"
	hostName     = "unknown"
	runnerPID    = os.Getpid()
	versionFlag  = flag.Bool("version", false, "Print version and exit")
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to config.yaml")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("%s\n", buildVersion())
		return
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fatalf("load config: %v", err)
	}

	buildVer = buildVersion()
	if h, err := os.Hostname(); err == nil {
		hostName = h
	}

	logger := setupLogger(cfg)

	printBanner(cfg, "(computed later)", buildVer)

	st, err := store.New(cfg.Storage.Path)
	if err != nil {
		fatalf("open store: %v", err)
	}
	defer func() {
		if err := st.Close(); err != nil {
			logger.Error("failed to close store", slog.String("err", err.Error()))
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	runner, err := app.Build(cfg, st, logger)
	if err != nil {
		fatalf("build runner: %v", err)
	}

	logger.Info("nostr-codex-runner starting")

	if err := runner.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		fatalf("runtime error: %v", err)
	}
}

func setupLogger(cfg *config.Config) *slog.Logger {
	level := slog.LevelInfo
	switch strings.ToLower(cfg.Logging.Level) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	writers := []io.Writer{os.Stdout}
	if cfg.Logging.File != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.Logging.File), 0o755); err != nil {
			fatalf("create log dir: %v", err)
		}
		f, err := os.OpenFile(cfg.Logging.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			fatalf("open log file: %v", err)
		}
		writers = append(writers, f)
	}
	logger := slog.New(slog.NewTextHandler(io.MultiWriter(writers...), &slog.HandlerOptions{Level: level}))
	return logger
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func buildVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "dev"
}

func printBanner(cfg *config.Config, pubKey string, version string) {
	fmt.Printf("\n==============================\n")
	fmt.Printf("nostr-codex-runner %s\n", version)
	fmt.Printf("pid %d • host %s • go %s\n", runnerPID, hostName, runtime.Version())
	fmt.Printf("config loaded\n")
	fmt.Printf("==============================\n\n")
}
