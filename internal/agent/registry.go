package agent

import (
	"fmt"
	"sync"

	"nostr-codex-runner/internal/core"
)

// Constructor builds an Agent from configuration.
type Constructor func(cfg any) (core.Agent, error)

var (
	registryMu sync.RWMutex
	registry   = make(map[string]Constructor)
)

func Register(kind string, ctor Constructor) error {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := registry[kind]; exists {
		return fmt.Errorf("agent type %s already registered", kind)
	}
	registry[kind] = ctor
	return nil
}

func MustRegister(kind string, ctor Constructor) {
	if err := Register(kind, ctor); err != nil {
		panic(err)
	}
}

func Build(kind string, cfg any) (core.Agent, error) {
	registryMu.RLock()
	ctor, ok := registry[kind]
	registryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown agent type %s", kind)
	}
	return ctor(cfg)
}

func RegisteredTypes() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]string, 0, len(registry))
	for k := range registry {
		out = append(out, k)
	}
	return out
}
