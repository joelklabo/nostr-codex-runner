package whatsapp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
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

func TestSendUsesTwilioAPI(t *testing.T) {
	var got url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/2010-04-01/Accounts/AC123/Messages.json" {
			t.Fatalf("path mismatch %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		got = r.Form
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	tr, err := New(Config{
		AccountSID: "AC123",
		AuthToken:  "token",
		FromNumber: "whatsapp:+15550001234",
		BaseURL:    srv.URL,
	}, nil)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	if err := tr.Send(context.Background(), core.OutboundMessage{Recipient: "+15550009999", Text: "hi"}); err != nil {
		t.Fatalf("send: %v", err)
	}
	if got == nil || got.Get("Body") != "hi" || got.Get("To") != "whatsapp:+15550009999" {
		t.Fatalf("form not captured: %+v", got)
	}
}

func TestSendReturnsErrorOnNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, "bad request")
	}))
	defer srv.Close()

	tr, err := New(Config{
		AccountSID: "AC123",
		AuthToken:  "token",
		FromNumber: "whatsapp:+15550001234",
		BaseURL:    srv.URL,
	}, nil)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	if err := tr.Send(context.Background(), core.OutboundMessage{Recipient: "+1", Text: "oops"}); err == nil {
		t.Fatalf("expected error for non-2xx")
	}
}

func TestVerifySignatureRejectsInvalid(t *testing.T) {
	cfg := Config{AccountSID: "AC", AuthToken: "token", FromNumber: "whatsapp:+1"}
	tr, _ := New(cfg, nil)
	form := url.Values{"From": {"whatsapp:+1"}}
	req, _ := http.NewRequest(http.MethodPost, "http://example.com/hook", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Twilio-Signature", "invalid")
	_ = req.ParseForm()
	if tr.verifySignature(req) {
		t.Fatalf("expected signature rejection")
	}
}

func TestSchemeFromHeader(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "http://x/h", nil)
	req.Header.Set("X-Forwarded-Proto", "https,http")
	if s := scheme(req); s != "https" {
		t.Fatalf("expected https, got %s", s)
	}
}

func TestContainsHelper(t *testing.T) {
	if !contains([]string{"a", "b"}, "b") {
		t.Fatalf("should contain")
	}
	if contains([]string{"a", "b"}, "c") {
		t.Fatalf("should not contain")
	}
}
