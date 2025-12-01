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
	"nostr-codex-runner/internal/metrics"
	"nostr-codex-runner/internal/store"
	"nostr-codex-runner/internal/wizard"
	"runtime"
	"runtime/debug"
)

var (
	buildVer  = "dev"
	hostName  = "unknown"
	runnerPID = os.Getpid()
)

const (
	envConfigNew    = "BUDDY_CONFIG"
	envConfigLegacy = "NCR_CONFIG"
)

func main() {
	subcmd, args := parseSubcommand(os.Args[1:])

	switch subcmd {
	case "version":
		fmt.Printf("%s\n", buildVersion())
		return
	case "help", "-h", "--help":
		usage()
		return
	case "wizard":
		if err := runWizard(args); err != nil {
			fatalf(err.Error())
		}
		return
	case "run":
		if err := runContext(context.Background(), args); err != nil {
			fatalf(err.Error())
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", subcmd)
		usage()
		os.Exit(2)
	}
}

func runContext(parent context.Context, args []string) error {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	configPath := fs.String("config", defaultConfigPath(), "Path to config.yaml (defaults: BUDDY_CONFIG env, ./config.yaml, ~/.config/buddy/config.yaml)")
	healthListen := fs.String("health-listen", "", "Optional health endpoint listen addr (e.g., 127.0.0.1:8081)")
	metricsListen := fs.String("metrics-listen", "", "Optional Prometheus metrics listen addr (e.g., 127.0.0.1:9090)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "unexpected arguments: %v\n\n", fs.Args())
		usage()
		return fmt.Errorf("unexpected arguments: %v", fs.Args())
	}

	ctx, stop := signal.NotifyContext(parent, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load(*configPath)
	if err != nil {
		return friendlyConfigErr(*configPath, err)
	}

	for _, w := range collectCompatWarnings(*configPath) {
		fmt.Fprintf(os.Stderr, "[warn] %s\n", w)
	}

	buildVer = buildVersion()
	if h, err := os.Hostname(); err == nil {
		hostName = h
	}

	logger := setupLogger(cfg)

	printBanner(cfg, "(computed later)", buildVer)

	st, err := store.New(cfg.Storage.Path)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer func() {
		if err := st.Close(); err != nil {
			logger.Error("failed to close store", slog.String("err", err.Error()))
		}
	}()

	runner, err := app.Build(cfg, st, logger)
	if err != nil {
		return fmt.Errorf("build runner: %w", err)
	}

	if *healthListen != "" {
		if _, err := health.Start(ctx, *healthListen, buildVer, logger); err != nil {
			return fmt.Errorf("start health: %w", err)
		}
	}
	if err := metrics.Start(ctx, *metricsListen, logger); err != nil {
		return fmt.Errorf("start metrics: %w", err)
	}

	logger.Info("nostr-codex-runner starting")

	if err := runner.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("runtime error: %w", err)
	}
	return nil
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
	if v := os.Getenv(envConfigNew); v != "" {
		return v
	}
	if v := os.Getenv(envConfigLegacy); v != "" {
		return v
	}
	// cwd config wins if present
	if fileExists("config.yaml") {
		return "config.yaml"
	}
	home, err := os.UserHomeDir()
	if err == nil {
		newPath := filepath.Join(home, ".config", "buddy", "config.yaml")
		if fileExists(newPath) {
			return newPath
		}
		legacyPath := filepath.Join(home, ".config", "nostr-codex-runner", "config.yaml")
		if fileExists(legacyPath) {
			return legacyPath
		}
		return newPath
	}
	return "config.yaml"
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  nostr-codex-runner run [-config path] [-health-listen addr] [-metrics-listen addr]\n")
	fmt.Fprintf(os.Stderr, "  nostr-codex-runner wizard [config-path]\n")
	fmt.Fprintf(os.Stderr, "  nostr-codex-runner version\n")
	fmt.Fprintf(os.Stderr, "  nostr-codex-runner help\n\n")
	fmt.Fprintf(os.Stderr, "Environment:\n")
	fmt.Fprintf(os.Stderr, "  %s\tdefault config path (overrides -config default)\n", envConfigNew)
	fmt.Fprintf(os.Stderr, "  %s\tlegacy default config path (deprecated)\n", envConfigLegacy)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func friendlyConfigErr(path string, err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("config not found at %s. Searched (in order): %s. Hint: run the wizard or copy config.example.yaml.", path, strings.Join(configSearchOrder(), ", "))
	}
	return fmt.Errorf("load config %s: %w", path, err)
}

func configSearchOrder() []string {
	paths := []string{}
	if v := os.Getenv(envConfigNew); v != "" {
		paths = append(paths, v+" (BUDDY_CONFIG)")
	}
	if v := os.Getenv(envConfigLegacy); v != "" {
		paths = append(paths, v+" (NCR_CONFIG legacy)")
	}
	paths = append(paths, "config.yaml (cwd)")
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".config", "buddy", "config.yaml"))
		paths = append(paths, filepath.Join(home, ".config", "nostr-codex-runner", "config.yaml")+" (legacy)")
	}
	return paths
}

func runWizard(args []string) error {
	var path string
	if len(args) > 0 {
		path = args[0]
	}
	cfgPath, err := wizard.Run(context.Background(), path, nil)
	if err != nil {
		return err
	}
	fmt.Printf("Config written to %s\n", cfgPath)
	fmt.Printf("Next: run `nostr-codex-runner run -config %s`\n", cfgPath)
	return nil
}

func collectCompatWarnings(configPath string) []string {
	var warnings []string
	if os.Getenv(envConfigLegacy) != "" {
		warnings = append(warnings, fmt.Sprintf("%s is deprecated; use %s instead", envConfigLegacy, envConfigNew))
	}
	if strings.Contains(configPath, ".config/nostr-codex-runner") {
		warnings = append(warnings, "config path uses legacy directory (.config/nostr-codex-runner); prefer ~/.config/buddy/config.yaml")
	}
	bin := filepath.Base(os.Args[0])
	if strings.Contains(bin, "nostr-codex-runner") {
		warnings = append(warnings, "binary name nostr-codex-runner is deprecated; future releases will use buddy/nostr-buddy")
	}
	return warnings
}
