package nostrclient

import (
	"context"
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
	pool    *nostr.SimplePool
	privKey string
	pubKey  string
	relays  []string
	store   *store.Store
	allowed map[string]struct{}

	secretMu sync.Mutex
	secrets  map[string][]byte

	seen *seenIDs
}

// New constructs a client pointing at the provided relays.
func New(privKey string, pubKey string, relays []string, allowedPubkeys []string, st *store.Store) *Client {
	allowed := make(map[string]struct{}, len(allowedPubkeys))
	for _, pk := range allowedPubkeys {
		allowed[strings.ToLower(pk)] = struct{}{}
	}
	return &Client{
		pool:    nostr.NewSimplePool(context.Background()),
		privKey: privKey,
		pubKey:  strings.ToLower(pubKey),
		relays:  relays,
		store:   st,
		allowed: allowed,
		secrets: make(map[string][]byte),
		seen:    newSeenIDs(),
	}
}

// Listen subscribes to encrypted DMs addressed to this runner and invokes handler for each new message.
func (c *Client) Listen(ctx context.Context, handler func(context.Context, IncomingMessage)) error {
	if c.pool == nil {
		return errors.New("nil pool")
	}

	filter := c.buildFilter()

	for {
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

				plaintext, err := nip04.Decrypt(evt.Content, secret)
				if err != nil {
					continue
				}

				_ = c.store.SaveCursor(sender, evt.CreatedAt.Time())

				go handler(ctx, IncomingMessage{Event: evt, SenderPubKey: sender, Plaintext: plaintext})
			}
		}
	resubscribe:
		continue
	}
}

func (c *Client) buildFilter() nostr.Filter {
	since := nostr.Timestamp(time.Now().Add(-2 * time.Hour).Unix())
	for pk := range c.allowed {
		if t, err := c.store.LastCursor(pk); err == nil && !t.IsZero() {
			ts := nostr.Timestamp(t.Unix())
			if ts < since {
				since = ts
			}
		}
	}
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
