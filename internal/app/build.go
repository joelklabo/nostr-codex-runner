package app

import (
	"fmt"
	"log/slog"
	"time"

	"nostr-codex-runner/internal/actions/fs"
	"nostr-codex-runner/internal/actions/shell"
	"nostr-codex-runner/internal/agents/codexcli"
	"nostr-codex-runner/internal/agents/echo"
	"nostr-codex-runner/internal/agents/http"
	"nostr-codex-runner/internal/config"
	"nostr-codex-runner/internal/core"
	"nostr-codex-runner/internal/store"
	tmock "nostr-codex-runner/internal/transports/mock"
	tnostr "nostr-codex-runner/internal/transports/nostr"
)

// Build constructs transports, agent, and actions from config.
func Build(cfg *config.Config, st *store.Store, logger *slog.Logger) (*core.Runner, error) {
	transports := make([]core.Transport, 0, len(cfg.Transports))
	for _, t := range cfg.Transports {
		switch t.Type {
		case "nostr":
			nt, err := tnostr.New(tnostr.Config{Relays: t.Relays, PrivateKey: t.PrivateKey, AllowedPubkeys: t.AllowedPubkeys}, st)
			if err != nil {
				return nil, err
			}
			transports = append(transports, nt)
		case "mock":
			transports = append(transports, tmock.New(t.ID))
		default:
			return nil, fmt.Errorf("unknown transport type %s", t.Type)
		}
	}

	var agent core.Agent
	switch cfg.Agent.Type {
	case "codexcli", "":
		agent = codexcli.New(codexcli.Config(cfg.Agent.Codex))
	case "echo":
		agent = echo.New()
	case "http":
		agent = http.New(http.Config{APIBase: cfg.Agent.Codex.Binary, Model: cfg.Agent.Codex.Profile}) // placeholder reuse fields
	default:
		return nil, fmt.Errorf("unknown agent type %s", cfg.Agent.Type)
	}

	actions := make([]core.Action, 0, len(cfg.Actions))
	for _, a := range cfg.Actions {
		switch a.Type {
		case "shell":
			actions = append(actions, shell.New(shell.Config{
				Workdir:        a.Workdir,
				Allowed:        a.Allowed,
				TimeoutSeconds: a.TimeoutSecs,
				MaxOutput:      a.MaxOutput,
			}))
		case "readfile":
			actions = append(actions, fs.NewReadFile(fs.Config{Roots: a.Roots, MaxBytes: a.MaxBytes}))
		case "writefile":
			actions = append(actions, fs.NewWriteFile(fs.Config{Roots: a.Roots, MaxBytes: a.MaxBytes, AllowWrite: a.AllowWrite}))
		default:
			return nil, fmt.Errorf("unknown action type %s", a.Type)
		}
	}

	r := core.NewRunner(transports, agent, actions, logger,
		core.WithAllowedSenders(cfg.Runner.AllowedPubkeys),
		core.WithStore(st),
		core.WithSessionTimeout(time.Duration(cfg.Runner.SessionTimeoutMins)*time.Minute),
		core.WithInitialPrompt(cfg.Runner.InitialPrompt),
		core.WithMaxReplyChars(cfg.Runner.MaxReplyChars),
	)
	return r, nil
}
