package whatsapp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"

	"nostr-codex-runner/internal/core"
)

func TestWebhookInbound(t *testing.T) {
	cfg := Config{
		AccountSID:     "AC123",
		AuthToken:      "token",
		FromNumber:     "whatsapp:+15550009999",
		Listen:         "127.0.0.1:18083",
		Path:           "/twilio/webhook",
		AllowedNumbers: []string{"15555550100"},
	}
	tr, err := New(cfg, nil)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	inbound := make(chan core.InboundMessage, 1)
	go func() {
		if err := tr.Start(ctx, inbound); err != nil && ctx.Err() == nil {
			t.Errorf("start: %v", err)
		}
	}()

	addr := tr.Addr()
	if addr == "" {
		// wait a moment
		time.Sleep(50 * time.Millisecond)
		addr = tr.Addr()
		if addr == "" {
			t.Fatalf("server not up")
		}
	}

	form := url.Values{}
	form.Set("From", "whatsapp:15555550100")
	form.Set("Body", "hello wa")
	form.Set("MessageSid", "SM123")

	rawURL := "http://" + addr + "/twilio/webhook"
	sig := signFor(rawURL, form, cfg.SignatureKey)

	req, _ := http.NewRequest(http.MethodPost, rawURL, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Twilio-Signature", sig)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}

	select {
	case m := <-inbound:
		if m.Text != "hello wa" || m.Sender != "15555550100" {
			t.Fatalf("bad message %+v", m)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("no message")
	}
}

// helper to get bound addr from transport by hitting listener path
// signFor builds the Twilio signature for tests.
func signFor(rawURL string, form url.Values, key string) string {
	if key == "" {
		key = "token"
	}
	var keys []string
	for k := range form {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var payload strings.Builder
	payload.WriteString(rawURL)
	for _, k := range keys {
		payload.WriteString(k)
		payload.WriteString(form.Get(k))
	}
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(payload.String()))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
