package main

import (
	"context"
	"encoding/json"
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

	"github.com/joelklabo/buddy/internal/app"
	"github.com/joelklabo/buddy/internal/assets"
	"github.com/joelklabo/buddy/internal/check"
	"github.com/joelklabo/buddy/internal/config"
	"github.com/joelklabo/buddy/internal/health"
	"github.com/joelklabo/buddy/internal/metrics"
	"github.com/joelklabo/buddy/internal/presets"
	"github.com/joelklabo/buddy/internal/store"
	"github.com/joelklabo/buddy/internal/wizard"
	"runtime"
	"runtime/debug"
)

var (
	buildVer  = "dev"
	hostName  = "unknown"
	runnerPID = os.Getpid()
)

const envConfigNew = "BUDDY_CONFIG"

func main() {
	subcmd, args := parseSubcommand(os.Args[1:])

	switch subcmd {
	case "version":
		fmt.Printf("%s\n", buildVersion())
		return
	case "help", "-h", "--help":
		printHelp(args)
		return
	case "wizard":
		if err := runWizard(args); err != nil {
			fatalf(err.Error())
		}
		return
	case "init-config":
		if err := runInitConfig(args); err != nil {
			fatalf(err.Error())
		}
		return
	case "presets":
		if err := runPresets(args); err != nil {
			fatalf(err.Error())
		}
		return
	case "check":
		if err := runCheck(args); err != nil {
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
	var positional string
	// Allow one positional argument for preset or config path.
	if fs.NArg() > 1 {
		fmt.Fprintf(os.Stderr, "too many arguments: %v\n\n", fs.Args())
		printHelp([]string{"run"})
		return fmt.Errorf("too many arguments: %v", fs.Args())
	}
	if fs.NArg() == 1 {
		positional = fs.Arg(0)
	}

	ctx, stop := signal.NotifyContext(parent, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, _, err := loadConfigWithPresets(*configPath, positional)
	if err != nil {
		return err
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

	logger.Info("buddy starting")

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
	fmt.Printf("buddy %s\n", version)
	fmt.Printf("pid %d • host %s • go %s\n", runnerPID, hostName, runtime.Version())
	fmt.Printf("config loaded\n")
	fmt.Printf("==============================\n\n")
}

func parseSubcommand(args []string) (string, []string) {
	if len(args) == 0 {
		return "run", args
	}
	first := args[0]
	switch first {
	case "presets", "wizard", "init-config", "check", "version", "help", "run":
		return first, args[1:]
	}
	if strings.HasPrefix(first, "-") {
		return "run", args
	}
	// Treat unknown-first-token as positional preset/config for run.
	return "run", args
}

func defaultConfigPath() string {
	if v := os.Getenv(envConfigNew); v != "" {
		return v
	}
	if v := os.Getenv(envConfigNew); v != "" {
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
		return newPath
	}
	return "config.yaml"
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: buddy <command> [args]\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  run <preset|config>       start the runner\n")
	fmt.Fprintf(os.Stderr, "  check <preset|config>     verify dependencies for a preset/config\n")
	fmt.Fprintf(os.Stderr, "  wizard [config-path]      guided setup; supports dry-run\n")
	fmt.Fprintf(os.Stderr, "  init-config [path]        write example config (default ./config.yaml)\n")
	fmt.Fprintf(os.Stderr, "  presets [name]            list built-in presets or show one\n")
	fmt.Fprintf(os.Stderr, "  version                   show version\n")
	fmt.Fprintf(os.Stderr, "  help [command]            show help\n\n")
	fmt.Fprintf(os.Stderr, "Env: %s (preferred)\n", envConfigNew)
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
	if v := os.Getenv(envConfigNew); v != "" {
		paths = append(paths, v+" (NCR_CONFIG legacy)")
	}
	paths = append(paths, "config.yaml (cwd)")
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".config", "buddy", "config.yaml"))
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
	fmt.Printf("Next: run `buddy run -config %s`\n", cfgPath)
	return nil
}

func loadConfigWithPresets(flagConfig string, positional string) (*config.Config, string, error) {
	// If positional provided, prefer it; otherwise fallback to flag/default path.
	if positional != "" {
		// If file exists, load as config.
		if fileExists(positional) {
			cfg, err := config.Load(positional)
			if err != nil {
				return nil, positional, friendlyConfigErr(positional, err)
			}
			return cfg, positional, nil
		}
		// Try preset
		if data, err := presets.Get(positional); err == nil {
			cfg, err := config.LoadBytes(data, ".")
			if err != nil {
				return nil, positional, fmt.Errorf("load preset %s: %w", positional, err)
			}
			return cfg, positional, nil
		}
		return nil, positional, fmt.Errorf("path or preset %q not found", positional)
	}

	cfg, err := config.Load(flagConfig)
	if err != nil {
		return nil, flagConfig, friendlyConfigErr(flagConfig, err)
	}
	return cfg, flagConfig, nil
}

func collectCompatWarnings(configPath string) []string { return nil }

func printHelp(args []string) {
	if len(args) == 0 {
		usage()
		return
	}
	switch args[0] {
	case "run":
		fmt.Println("buddy run <preset|config> - start the runner using a preset or YAML config")
		fmt.Println("Flags:")
		fmt.Println("  -config <path>          config file path (default search: argv, ./config.yaml, ~/.config/buddy/config.yaml)")
		fmt.Println("  -health-listen <addr>   optional health endpoint (e.g., 127.0.0.1:8081)")
		fmt.Println("  -metrics-listen <addr>  optional Prometheus metrics endpoint")
		fmt.Println("Examples:")
		fmt.Println("  buddy run mock-echo                 # offline smoke test")
		fmt.Println("  buddy run claude-dm                 # Nostr → Claude/OpenAI HTTP")
		fmt.Println("  buddy run copilot-shell             # Copilot + shell (trusted operators only)")
		fmt.Println("  buddy run path/to/config.yaml       # load your own YAML")
	case "wizard":
		fmt.Println("buddy wizard [config-path] - guided setup; writes config or dry-runs")
		fmt.Println("Prompts for relays/keys/allowed pubkeys, agent choice, actions.")
	case "init-config":
		fmt.Println("buddy init-config [path] - write config.example.yaml to path (default ./config.yaml) if missing")
	case "presets":
		fmt.Println("buddy presets [name] - list built-in presets or show one.")
		fmt.Println("Examples:")
		fmt.Println("  buddy presets                         # list all presets")
		fmt.Println("  buddy presets claude-dm               # show summary and how to run")
		fmt.Println("  buddy presets claude-dm --yaml        # print the YAML")
	case "check":
		fmt.Println("buddy check <preset|config> - verify declared dependencies")
		fmt.Println("Examples:")
		fmt.Println("  buddy check mock-echo")
		fmt.Println("  buddy check copilot-shell")
		fmt.Println("  buddy check path/to/config.yaml")
		fmt.Println("Flags:")
		fmt.Println("  -config <path>          config file path (default search: argv, ./config.yaml, ~/.config/buddy/config.yaml)")
		fmt.Println("  -json                   output JSON report")
	case "version":
		fmt.Println("buddy version - print version")
	default:
		usage()
	}
}

func runPresets(args []string) error {
	if len(args) == 0 {
		for name, desc := range presets.List() {
			fmt.Printf("%-22s %s\n", name, desc)
		}
		return nil
	}
	name := args[0]
	if len(args) == 2 && args[1] == "--yaml" {
		data, err := presets.Get(name)
		if err != nil {
			return err
		}
		fmt.Print(string(data))
		return nil
	}
	descMap := presets.List()
	if _, ok := descMap[name]; !ok {
		return fmt.Errorf("unknown preset %s", name)
	}
	fmt.Printf("Preset: %s\n", name)
	fmt.Printf("Description: %s\n", descMap[name])
	fmt.Printf("Run it:\n  buddy run %s\n\n", name)
	fmt.Println("View YAML: buddy presets", name, "--yaml")
	return nil
}

func runCheck(args []string) error {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	configPath := fs.String("config", defaultConfigPath(), "Path to config.yaml or preset name")
	jsonOut := fs.Bool("json", false, "Output JSON report")
	if err := fs.Parse(args); err != nil {
		return err
	}
	var positional string
	if fs.NArg() > 1 {
		fmt.Fprintf(os.Stderr, "unexpected arguments: %v\n\n", fs.Args())
		usage()
		return fmt.Errorf("unexpected arguments: %v", fs.Args())
	}
	if fs.NArg() == 1 {
		positional = fs.Arg(0)
	}

	cfg, presetName, err := loadConfigWithPresets(*configPath, positional)
	if err != nil {
		return err
	}
	deps := check.AggregateDeps(cfg, presetName, presets.PresetDeps())
	if len(deps) == 0 {
		fmt.Println("No dependencies declared; nothing to check.")
		return nil
	}

	checkers := map[string]check.Checker{
		"binary": check.BinaryChecker{},
		"env":    check.EnvChecker{},
		"file":   check.FileChecker{},
		"url":    check.URLChecker{},
		"port":   check.PortChecker{},
	}

	results := make([]check.Result, 0, len(deps))
	for _, d := range deps {
		input := check.DepInput{
			Name:        d.Name,
			Type:        d.Type,
			Version:     d.Version,
			Optional:    d.Optional,
			Description: d.Description,
			Hint:        d.Hint,
		}
		chk, ok := checkers[d.Type]
		if !ok {
			results = append(results, check.Result{
				Name:     d.Name,
				Type:     d.Type,
				Status:   "WARN",
				Details:  "unsupported dep type",
				Optional: true,
			})
			continue
		}
		results = append(results, chk.Check(input))
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	fmt.Println("Dependency check")
	if presetName != "" {
		fmt.Printf("Preset: %s\n", presetName)
	}
	fmt.Println()

	missing := 0
	for _, r := range results {
		icon := "✅"
		switch r.Status {
		case "WARN":
			icon = "⚠️ "
		case "MISSING":
			icon = "❌"
			if !r.Optional {
				missing++
			}
		}
		fmt.Printf("%s %-6s %-16s %s\n", icon, r.Type, r.Name, r.Details)
	}

	if missing > 0 {
		return fmt.Errorf("%d required dependencies missing", missing)
	}
	return nil
}

func runInitConfig(args []string) error {
	path := "config.yaml"
	if len(args) > 0 {
		path = args[0]
	}
	if fileExists(path) {
		return fmt.Errorf("%s already exists", path)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("make dir: %w", err)
	}
	if err := os.WriteFile(path, assets.ConfigExample, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	fmt.Printf("Wrote example config to %s\n", path)
	return nil
}
