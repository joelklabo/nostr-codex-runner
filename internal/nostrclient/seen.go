package nostrclient

import "sync/atomic"

type seenIDs struct {
	// map of event ID string to a dummy value; we use atomic.Value to avoid locking overhead.
	v atomic.Value // stores map[string]struct{}
}

func newSeenIDs() *seenIDs {
	s := &seenIDs{}
	s.v.Store(make(map[string]struct{}))
	return s
}

func (s *seenIDs) Seen(id string) bool {
	m := s.v.Load().(map[string]struct{})
	_, ok := m[id]
	if ok {
		return true
	}
	// copy-on-write to keep things simple
	m2 := make(map[string]struct{}, len(m)+1)
	for k, v := range m {
		m2[k] = v
	}
	m2[id] = struct{}{}
	s.v.Store(m2)
	return false
}
