package nostrclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"nostr-codex-runner/internal/store"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
)

// IncomingMessage is a decrypted DM sent to the runner.
type IncomingMessage struct {
	Event        *nostr.Event
	SenderPubKey string
	Plaintext    string
}

// Client wraps Nostr connectivity and send/receive helpers.
type Client struct {
	pool    Pool
	privKey string
	pubKey  string
	relays  []string
	store   store.StoreAPI
	allowed map[string]struct{}

	secretMu sync.Mutex
	secrets  map[string][]byte

	seen *seenIDs

	lastMu    sync.Mutex
	lastMsg   map[string]lastSeen
	windowDur time.Duration

	senderMu    sync.Mutex
	senderLocks map[string]*sync.Mutex
	msgWindow   time.Duration
}

type lastSeen struct {
	text string
	ts   time.Time
}

// New constructs a client pointing at the provided relays.
func New(privKey string, pubKey string, relays []string, allowedPubkeys []string, st store.StoreAPI) *Client {
	return NewWithPool(privKey, pubKey, relays, allowedPubkeys, st, nostr.NewSimplePool(context.Background()))
}

// NewWithPool allows injecting a custom pool (for tests).
func NewWithPool(privKey string, pubKey string, relays []string, allowedPubkeys []string, st store.StoreAPI, pool Pool) *Client {
	allowed := make(map[string]struct{}, len(allowedPubkeys))
	for _, pk := range allowedPubkeys {
		allowed[strings.ToLower(pk)] = struct{}{}
	}
	return &Client{
		pool:        pool,
		privKey:     privKey,
		pubKey:      strings.ToLower(pubKey),
		relays:      relays,
		store:       st,
		allowed:     allowed,
		secrets:     make(map[string][]byte),
		seen:        newSeenIDs(),
		lastMsg:     make(map[string]lastSeen),
		windowDur:   8 * time.Second,
		senderLocks: make(map[string]*sync.Mutex),
		msgWindow:   30 * time.Second,
	}
}

// Listen subscribes to encrypted DMs addressed to this runner and invokes handler for each new message.
func (c *Client) Listen(ctx context.Context, handler func(context.Context, IncomingMessage)) error {
	if c.pool == nil {
		return errors.New("nil pool")
	}

	for {
		filter := c.buildFilter()
		events := c.pool.SubscribeMany(ctx, c.relays, filter)
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case ie, ok := <-events:
				if !ok {
					time.Sleep(2 * time.Second)
					goto resubscribe
				}
				evt := ie.Event
				if evt == nil {
					continue
				}

				if c.seen.Seen(evt.ID) {
					continue
				}
				already, err := c.store.AlreadyProcessed(evt.ID)
				if err != nil {
					continue
				}
				if already {
					continue
				}

				sender := strings.ToLower(evt.PubKey)
				if _, ok := c.allowed[sender]; !ok {
					continue
				}

				secret, err := c.sharedSecret(sender)
				if err != nil {
					continue
				}

				lock := c.senderLock(sender)
				go func(e *nostr.Event, s string, sec []byte) {
					lock.Lock()
					defer lock.Unlock()

					dec, err := nip04.Decrypt(e.Content, sec)
					if err != nil {
						return
					}

					if seen, err := c.store.RecentMessageSeen(s, dec, c.msgWindow); err == nil && seen {
						return
					}

					if c.isReplay(s, dec, e.CreatedAt.Time()) {
						return
					}

					_ = c.store.SaveCursor(s, e.CreatedAt.Time())

					handler(ctx, IncomingMessage{Event: e, SenderPubKey: s, Plaintext: dec})
				}(evt, sender, secret)
			}
		}
	resubscribe:
		continue
	}
}

func (c *Client) buildFilter() nostr.Filter {
	since := c.lastCursorMax()
	return nostr.Filter{
		Kinds:   []int{nostr.KindEncryptedDirectMessage},
		Authors: c.allowedList(),
		Since:   &since,
		Tags:    nostr.TagMap{"p": []string{c.pubKey}},
	}
}

// SendReply DM's a message back to the sender.
func (c *Client) SendReply(ctx context.Context, toPubKey string, message string) error {
	secret, err := c.sharedSecret(toPubKey)
	if err != nil {
		return err
	}

	enc, err := nip04.Encrypt(message, secret)
	if err != nil {
		return fmt.Errorf("encrypt DM: %w", err)
	}

	ev := nostr.Event{
		PubKey:    c.pubKey,
		CreatedAt: nostr.Now(),
		Kind:      nostr.KindEncryptedDirectMessage,
		Tags:      nostr.Tags{nostr.Tag{"p", toPubKey}},
		Content:   enc,
	}
	if err := ev.Sign(c.privKey); err != nil {
		return fmt.Errorf("sign DM: %w", err)
	}

	results := c.pool.PublishMany(ctx, c.relays, ev)
	var firstErr error
	for res := range results {
		if res.Error != nil && firstErr == nil {
			firstErr = res.Error
		}
	}
	return firstErr
}

// PublishProfile broadcasts the runner's metadata (name, picture) to configured relays.
func (c *Client) PublishProfile(ctx context.Context, name, picture string) error {
	meta := make(map[string]string, 2)
	if strings.TrimSpace(name) != "" {
		meta["name"] = name
	}
	if strings.TrimSpace(picture) != "" {
		meta["picture"] = picture
	}
	if len(meta) == 0 {
		return nil
	}

	content, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshal profile: %w", err)
	}

	ev := nostr.Event{
		PubKey:    c.pubKey,
		CreatedAt: nostr.Now(),
		Kind:      nostr.KindProfileMetadata,
		Content:   string(content),
	}
	if err := ev.Sign(c.privKey); err != nil {
		return fmt.Errorf("sign profile: %w", err)
	}

	results := c.pool.PublishMany(ctx, c.relays, ev)
	var firstErr error
	for res := range results {
		if res.Error != nil && firstErr == nil {
			firstErr = res.Error
		}
	}
	return firstErr
}

func (c *Client) sharedSecret(peerPub string) ([]byte, error) {
	peerPub = strings.ToLower(peerPub)
	c.secretMu.Lock()
	if key, ok := c.secrets[peerPub]; ok {
		c.secretMu.Unlock()
		return key, nil
	}
	c.secretMu.Unlock()

	key, err := nip04.ComputeSharedSecret(peerPub, c.privKey)
	if err != nil {
		return nil, fmt.Errorf("compute shared secret: %w", err)
	}

	c.secretMu.Lock()
	c.secrets[peerPub] = key
	c.secretMu.Unlock()
	return key, nil
}

func (c *Client) allowedList() []string {
	res := make([]string, 0, len(c.allowed))
	for pk := range c.allowed {
		res = append(res, pk)
	}
	return res
}

// lastCursorMax returns the most recent cursor across allowed senders,
// or defaults to now-30s to avoid replaying old messages. Rewinds 5s to catch in-flight events.
func (c *Client) lastCursorMax() nostr.Timestamp {
	latest := time.Now().Add(-30 * time.Second)
	for pk := range c.allowed {
		if t, err := c.store.LastCursor(pk); err == nil && !t.IsZero() && t.After(latest) {
			latest = t
		}
	}
	latest = latest.Add(-5 * time.Second)
	return nostr.Timestamp(latest.Unix())
}

// isReplay drops identical plaintext from same sender within windowDur.
func (c *Client) isReplay(sender, plaintext string, at time.Time) bool {
	c.lastMu.Lock()
	defer c.lastMu.Unlock()
	if c.windowDur == 0 {
		c.windowDur = 8 * time.Second
	}
	ls, ok := c.lastMsg[sender]
	if ok && ls.text == plaintext && at.Sub(ls.ts) < c.windowDur {
		return true
	}
	c.lastMsg[sender] = lastSeen{text: plaintext, ts: at}
	return false
}

// senderLock returns a per-sender mutex to serialize handling of messages from the same pubkey.
func (c *Client) senderLock(sender string) *sync.Mutex {
	c.senderMu.Lock()
	defer c.senderMu.Unlock()
	if m, ok := c.senderLocks[sender]; ok {
		return m
	}
	m := &sync.Mutex{}
	c.senderLocks[sender] = m
	return m
}
