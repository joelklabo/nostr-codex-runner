package transport

import (
	"fmt"
	"sync"

	"nostr-codex-runner/internal/core"
)

// Constructor builds a Transport from a config object (type-specific).
type Constructor func(cfg any) (core.Transport, error)

var (
	registryMu sync.RWMutex
	registry   = make(map[string]Constructor)
)

// Register adds a constructor for a transport type.
func Register(kind string, ctor Constructor) error {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := registry[kind]; exists {
		return fmt.Errorf("transport type %s already registered", kind)
	}
	registry[kind] = ctor
	return nil
}

// MustRegister panics on error; intended for init() in transport packages.
func MustRegister(kind string, ctor Constructor) {
	if err := Register(kind, ctor); err != nil {
		panic(err)
	}
}

// Build constructs a transport of the given type.
func Build(kind string, cfg any) (core.Transport, error) {
	registryMu.RLock()
	ctor, ok := registry[kind]
	registryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown transport type %s", kind)
	}
	return ctor(cfg)
}

// RegisteredTypes returns a snapshot of registered transport kinds.
func RegisteredTypes() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]string, 0, len(registry))
	for k := range registry {
		out = append(out, k)
	}
	return out
}
