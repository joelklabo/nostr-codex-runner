// Package whatsapp implements a Twilio WhatsApp transport modeled after warelay.
package whatsapp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"nostr-codex-runner/internal/core"
	transport "nostr-codex-runner/internal/transports"
)

// Config for Twilio WhatsApp.
type Config struct {
	ID             string   `json:"id"`
	AccountSID     string   `json:"account_sid"`
	AuthToken      string   `json:"auth_token"`
	FromNumber     string   `json:"from_number"` // e.g. "whatsapp:+1234567890"
	Listen         string   `json:"listen"`      // ":8083"
	Path           string   `json:"path"`        // "/twilio/webhook"
	AllowedNumbers []string `json:"allowed_numbers"`
	SignatureKey   string   `json:"signature_key"` // optional; falls back to AuthToken
	BaseURL        string   `json:"base_url"`      // optional Twilio API base override for tests
}

type Transport struct {
	cfg Config
	log *slog.Logger

	addrMu sync.RWMutex
	addr   string
}

func New(cfg Config, logger *slog.Logger) (*Transport, error) {
	if cfg.ID == "" {
		cfg.ID = "whatsapp"
	}
	if cfg.AccountSID == "" || cfg.AuthToken == "" || cfg.FromNumber == "" {
		return nil, errors.New("whatsapp: account_sid, auth_token, from_number required")
	}
	if cfg.Listen == "" {
		cfg.Listen = ":8083"
	}
	if cfg.Path == "" {
		cfg.Path = "/twilio/webhook"
	}
	if cfg.SignatureKey == "" {
		cfg.SignatureKey = cfg.AuthToken
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.twilio.com"
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Transport{cfg: cfg, log: logger.With("transport", "whatsapp")}, nil
}

func (t *Transport) ID() string { return t.cfg.ID }

func (t *Transport) Start(ctx context.Context, inbound chan<- core.InboundMessage) error {
	mux := http.NewServeMux()
	mux.HandleFunc(t.cfg.Path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if !t.verifySignature(r) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		from := strings.TrimPrefix(r.Form.Get("From"), "whatsapp:")
		body := r.Form.Get("Body")
		msgID := r.Form.Get("MessageSid")
		if len(t.cfg.AllowedNumbers) > 0 && !contains(t.cfg.AllowedNumbers, from) {
			t.log.Warn("rejecting sender not in allowlist", "from", from)
			w.WriteHeader(http.StatusOK)
			return
		}
		im := core.InboundMessage{
			Transport: t.ID(),
			Sender:    from,
			Text:      body,
			ThreadID:  msgID,
		}
		select {
		case inbound <- im:
		case <-ctx.Done():
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	ln, err := net.Listen("tcp", t.cfg.Listen)
	if err != nil {
		return err
	}
	t.addrMu.Lock()
	t.addr = ln.Addr().String()
	t.addrMu.Unlock()
	errCh := make(chan error, 1)
	go func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		_ = srv.Shutdown(context.Background())
		return nil
	case err := <-errCh:
		return err
	}
}

// Addr returns the listen address (for tests).
func (t *Transport) Addr() string {
	t.addrMu.RLock()
	defer t.addrMu.RUnlock()
	return t.addr
}

func (t *Transport) Send(ctx context.Context, msg core.OutboundMessage) error {
	form := url.Values{}
	form.Set("To", "whatsapp:"+msg.Recipient)
	form.Set("From", t.cfg.FromNumber)
	form.Set("Body", msg.Text)
	api := fmt.Sprintf("%s/2010-04-01/Accounts/%s/Messages.json", strings.TrimRight(t.cfg.BaseURL, "/"), t.cfg.AccountSID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, api, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(t.cfg.AccountSID, t.cfg.AuthToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("twilio send failed: %s", strings.TrimSpace(string(b)))
	}
	return nil
}

func (t *Transport) verifySignature(r *http.Request) bool {
	sig := r.Header.Get("X-Twilio-Signature")
	if sig == "" {
		return false
	}
	// Twilio signature: base64(HMAC-SHA256(token, url + sorted params))
	rawURL := fmt.Sprintf("%s://%s%s", scheme(r), r.Host, r.URL.Path)
	params := r.PostForm
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var payload strings.Builder
	payload.WriteString(rawURL)
	for _, k := range keys {
		payload.WriteString(k)
		payload.WriteString(params.Get(k))
	}
	mac := hmac.New(sha256.New, []byte(t.cfg.SignatureKey))
	mac.Write([]byte(payload.String()))
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(sig), []byte(expected))
}

func scheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	if h := r.Header.Get("X-Forwarded-Proto"); h != "" {
		return strings.Split(h, ",")[0]
	}
	return "http"
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func init() {
	transport.MustRegister("whatsapp", func(cfg any) (core.Transport, error) {
		c, ok := cfg.(Config)
		if !ok {
			return nil, fmt.Errorf("whatsapp: invalid config type %T", cfg)
		}
		return New(c, nil)
	})
}
