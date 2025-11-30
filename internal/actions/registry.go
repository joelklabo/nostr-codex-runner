package action

import (
	"fmt"
	"sync"

	"nostr-codex-runner/internal/core"
)

// Constructor builds an Action from config.
type Constructor func(cfg any) (core.Action, error)

var (
	registryMu sync.RWMutex
	registry   = make(map[string]Constructor)
)

func Register(name string, ctor Constructor) error {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, ok := registry[name]; ok {
		return fmt.Errorf("action %s already registered", name)
	}
	registry[name] = ctor
	return nil
}

func MustRegister(name string, ctor Constructor) {
	if err := Register(name, ctor); err != nil {
		panic(err)
	}
}

func Build(name string, cfg any) (core.Action, error) {
	registryMu.RLock()
	ctor, ok := registry[name]
	registryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown action %s", name)
	}
	return ctor(cfg)
}

func RegisteredNames() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]string, 0, len(registry))
	for k := range registry {
		out = append(out, k)
	}
	return out
}
