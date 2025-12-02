package app

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/joelklabo/buddy/internal/actions/fs"
	"github.com/joelklabo/buddy/internal/actions/shell"
	"github.com/joelklabo/buddy/internal/agents/codexcli"
	"github.com/joelklabo/buddy/internal/agents/copilotcli"
	"github.com/joelklabo/buddy/internal/agents/echo"
	"github.com/joelklabo/buddy/internal/agents/http"
	"github.com/joelklabo/buddy/internal/config"
	"github.com/joelklabo/buddy/internal/core"
	"github.com/joelklabo/buddy/internal/store"
	imap "github.com/joelklabo/buddy/internal/transports/email/imap"
	memg "github.com/joelklabo/buddy/internal/transports/email/mailgun"
	tmock "github.com/joelklabo/buddy/internal/transports/mock"
	tnostr "github.com/joelklabo/buddy/internal/transports/nostr"
	twa "github.com/joelklabo/buddy/internal/transports/whatsapp"
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
		case "email":
			mode, _ := t.Config["mode"].(string)
			if mode == "" {
				mode = "mailgun"
			}
			switch mode {
			case "mailgun":
				var mcfg memg.Config
				if err := decodeMap(t.Config, &mcfg); err != nil {
					return nil, fmt.Errorf("decode mailgun config: %w", err)
				}
				if mcfg.ID == "" {
					mcfg.ID = t.ID
				}
				mt, err := memg.New(mcfg)
				if err != nil {
					return nil, err
				}
				transports = append(transports, mt)
			case "imap":
				var icfg imap.Config
				if err := decodeMap(t.Config, &icfg); err != nil {
					return nil, fmt.Errorf("decode imap config: %w", err)
				}
				if icfg.ID == "" {
					icfg.ID = t.ID
				}
				it, err := imap.New(icfg)
				if err != nil {
					return nil, err
				}
				transports = append(transports, it)
			default:
				return nil, fmt.Errorf("unknown email mode %s", mode)
			}
		case "whatsapp":
			var wcfg twa.Config
			if err := decodeMap(t.Config, &wcfg); err != nil {
				return nil, fmt.Errorf("decode whatsapp config: %w", err)
			}
			// fallback to ID if provided at top-level
			if wcfg.ID == "" {
				wcfg.ID = t.ID
			}
			wt, err := twa.New(wcfg, logger)
			if err != nil {
				return nil, err
			}
			transports = append(transports, wt)
		default:
			return nil, fmt.Errorf("unknown transport type %s", t.Type)
		}
	}

	agentCfg := cfg.Agent.Config
	var agent core.Agent
	switch cfg.Agent.Type {
	case "codexcli", "":
		agent = codexcli.New(codexcli.Config(agentCfg))
	case "echo":
		agent = echo.New()
	case "http":
		agent = http.New(http.Config{APIBase: agentCfg.Binary, Model: agentCfg.Profile}) // placeholder reuse fields
	case "copilotcli":
		agent = copilotcli.New(copilotcli.Config{
			Binary:         agentCfg.Binary,
			WorkingDir:     agentCfg.WorkingDir,
			TimeoutSeconds: agentCfg.TimeoutSeconds,
			AllowAllTools:  false,
			ExtraArgs:      agentCfg.ExtraArgs,
		})
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

// decodeMap marshals a generic map into a typed struct via JSON.
func decodeMap(m map[string]any, out any) error {
	if len(m) == 0 {
		return nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}
