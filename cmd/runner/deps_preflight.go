package main

import (
	"fmt"

	"github.com/joelklabo/buddy/internal/check"
	"github.com/joelklabo/buddy/internal/config"
	"github.com/joelklabo/buddy/internal/presets"
)

// runDepPreflight executes dependency checks for the current config/preset.
// It returns an error if any required dependency is missing.
func runDepPreflight(cfg *config.Config, presetName string) error {
	deps := check.AggregateDeps(cfg, presetName, presets.PresetDeps())
	if len(deps) == 0 {
		return nil
	}
	checkers := map[string]check.Checker{
		"binary":   check.BinaryChecker{},
		"env":      check.EnvChecker{},
		"file":     check.FileChecker{},
		"url":      check.URLChecker{},
		"port":     check.PortChecker{},
		"relay":    check.RelayChecker{},
		"dirwrite": check.DirWriteChecker{},
	}
	missing := 0
	for _, d := range deps {
		chk, ok := checkers[d.Type]
		if !ok {
			continue
		}
		res := chk.Check(check.DepInput{
			Name:     d.Name,
			Type:     d.Type,
			Version:  d.Version,
			Optional: d.Optional,
			Hint:     d.Hint,
		})
		if res.Status == "MISSING" {
			missing++
			fmt.Printf("❌ %s (%s) — %s\n", res.Name, res.Type, res.Details)
		} else if res.Status == "WARN" {
			fmt.Printf("⚠️  %s (%s) — %s\n", res.Name, res.Type, res.Details)
		}
	}
	if missing > 0 {
		return fmt.Errorf("%d required dependencies missing; rerun with -skip-check to bypass", missing)
	}
	return nil
}
