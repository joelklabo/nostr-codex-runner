package nostr

import (
	"testing"

	"nostr-codex-runner/internal/store"

	"github.com/nbd-wtf/go-nostr"
)

func TestNewMissingKey(t *testing.T) {
	if _, err := New(Config{}, nil); err == nil {
		t.Fatalf("expected error for missing key")
	}
}

func TestNewAndID(t *testing.T) {
	priv := nostr.GeneratePrivateKey()
	st, _ := store.New(t.TempDir() + "/state.db")
	defer st.Close()
	tr, err := New(Config{PrivateKey: priv}, st)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if tr.ID() != "nostr" {
		t.Fatalf("id mismatch")
	}
}
