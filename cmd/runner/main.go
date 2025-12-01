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
	"strings"
	"syscall"

	"nostr-codex-runner/internal/app"
	"nostr-codex-runner/internal/config"
	"nostr-codex-runner/internal/health"
	"nostr-codex-runner/internal/store"
	"runtime"
	"runtime/debug"
)

var (
	buildVer  = "dev"
	hostName  = "unknown"
	runnerPID = os.Getpid()
)

const envConfig = "NCR_CONFIG"

func main() {
	subcmd, args := parseSubcommand(os.Args[1:])

	switch subcmd {
	case "version":
		fmt.Printf("%s\n", buildVersion())
		return
	case "run":
		run(args)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", subcmd)
		usage()
		os.Exit(2)
	}
}

func run(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	configPath := fs.String("config", defaultConfigPath(), "Path to config.yaml")
	healthListen := fs.String("health-listen", "", "Optional health endpoint listen addr (e.g., 127.0.0.1:8081)")
	if err := fs.Parse(args); err != nil {
		fatalf(err.Error())
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "unexpected arguments: %v\n\n", fs.Args())
		usage()
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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

	runner, err := app.Build(cfg, st, logger)
	if err != nil {
		fatalf("build runner: %v", err)
	}

	if *healthListen != "" {
		if _, err := health.Start(ctx, *healthListen, buildVer, logger); err != nil {
			fatalf("start health: %v", err)
		}
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
		if err := os.MkdirAll(filepath.Dir(cfg.Logging.File), 0o750); err != nil {
			fatalf("create log dir: %v", err)
		}
		f, err := os.OpenFile(cfg.Logging.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			fatalf("open log file: %v", err)
		}
		writers = append(writers, f)
	}
	handlerOpts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler = slog.NewTextHandler(io.MultiWriter(writers...), handlerOpts)
	if strings.ToLower(cfg.Logging.Format) == "json" {
		handler = slog.NewJSONHandler(io.MultiWriter(writers...), handlerOpts)
	}
	logger := slog.New(handler)
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

func parseSubcommand(args []string) (string, []string) {
	if len(args) == 0 {
		return "run", args
	}
	first := args[0]
	if strings.HasPrefix(first, "-") {
		return "run", args
	}
	return first, args[1:]
}

func defaultConfigPath() string {
	if v := os.Getenv(envConfig); v != "" {
		return v
	}
	return "config.yaml"
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  nostr-codex-runner run [-config path] [-health-listen addr]\n")
	fmt.Fprintf(os.Stderr, "  nostr-codex-runner version\n\n")
	fmt.Fprintf(os.Stderr, "Environment:\n")
	fmt.Fprintf(os.Stderr, "  %s\tdefault config path (overrides -config default)\n", envConfig)
}
