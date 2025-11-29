package store

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	bucketActive    = []byte("active_sessions")
	bucketCursor    = []byte("cursors")
	bucketProcessed = []byte("processed")
	bucketMessages  = []byte("messages")
)

// SessionState represents the current Codex session for a sender.
type SessionState struct {
	SessionID string    `json:"session_id"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Store wraps a BoltDB instance for small, durable state.
type Store struct {
	db *bolt.DB
}

// New opens (or creates) the database at the given path.
func New(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(bucketActive); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(bucketCursor); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(bucketProcessed); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(bucketMessages); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

// Close releases the underlying DB handle.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

// SaveActive stores the active session for a given sender pubkey.
func (s *Store) SaveActive(pubkey, sessionID string) error {
	st := SessionState{SessionID: sessionID, UpdatedAt: time.Now().UTC()}
	data, err := json.Marshal(st)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketActive).Put([]byte(pubkey), data)
	})
}

// ClearActive removes the active session for a sender.
func (s *Store) ClearActive(pubkey string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketActive).Delete([]byte(pubkey))
	})
}

// Active returns the session state for a sender, if present.
func (s *Store) Active(pubkey string) (SessionState, bool, error) {
	var st SessionState
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketActive)
		data := b.Get([]byte(pubkey))
		if data == nil {
			return nil
		}
		if err := json.Unmarshal(data, &st); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return st, false, err
	}
	if st.SessionID == "" {
		return st, false, nil
	}
	return st, true, nil
}

// LastCursor returns the last event timestamp we processed for this sender.
func (s *Store) LastCursor(pubkey string) (time.Time, error) {
	var ts time.Time
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketCursor)
		v := b.Get([]byte(pubkey))
		if v == nil {
			ts = time.Time{}
			return nil
		}
		// timestamps are stored as RFC3339
		parsed, err := time.Parse(time.RFC3339Nano, string(v))
		if err != nil {
			return err
		}
		ts = parsed
		return nil
	})
	return ts, err
}

// SaveCursor persists the last event timestamp for a sender.
func (s *Store) SaveCursor(pubkey string, t time.Time) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketCursor).Put([]byte(pubkey), []byte(t.UTC().Format(time.RFC3339Nano)))
	})
}

// AlreadyProcessed checks if we've handled an event ID; if not, it marks it processed.
func (s *Store) AlreadyProcessed(id string) (bool, error) {
	if id == "" {
		return false, errors.New("empty event id")
	}
	var existed bool
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketProcessed)
		if v := b.Get([]byte(id)); v != nil {
			existed = true
			return nil
		}
		return b.Put([]byte(id), []byte{1})
	})
	return existed, err
}

// RecentMessageSeen returns true if the same sender/plaintext was seen within the window.
// It also records the current occurrence.
func (s *Store) RecentMessageSeen(sender, plaintext string, window time.Duration) (bool, error) {
	if window <= 0 {
		window = 30 * time.Second
	}
	sender = strings.ToLower(strings.TrimSpace(sender))
	body := strings.TrimSpace(plaintext)
	h := sha256.Sum256([]byte(body))
	key := sender + ":" + hex.EncodeToString(h[:])
	now := time.Now().UTC()

	var seen bool
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketMessages)
		v := b.Get([]byte(key))
		if v != nil {
			if ts, err := time.Parse(time.RFC3339Nano, string(v)); err == nil {
				if now.Sub(ts) < window {
					seen = true
				}
			}
		}
		return b.Put([]byte(key), []byte(now.Format(time.RFC3339Nano)))
	})
	return seen, err
}
