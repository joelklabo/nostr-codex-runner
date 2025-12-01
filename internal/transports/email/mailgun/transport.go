package mailgun

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/joelklabo/buddy/internal/core"
)

// Transport implements core.Transport backed by Mailgun inbound webhooks for inbound
// and Mailgun send API for outbound.
type Transport struct {
	cfg    Config
	client *Client
}

// New builds a Transport from Config (call Defaults/Validate upstream).
func New(cfg Config) (*Transport, error) {
	cfg.Defaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Transport{
		cfg:    cfg,
		client: NewClient(cfg.BaseURL, cfg.Domain, cfg.APIKey, cfg.Timeout),
	}, nil
}

func (t *Transport) ID() string {
	if t.cfg.ID != "" {
		return t.cfg.ID
	}
	return "email-mailgun"
}

// Start registers an HTTP handler; caller must wire the route. We keep a handler factory
// so Start stays non-blocking for compatibility with other transports.
func (t *Transport) Start(ctx context.Context, inbound chan<- core.InboundMessage) error {
	// non-blocking: caller should mount Handler() on their mux
	<-ctx.Done()
	return ctx.Err()
}

// Handler returns an http.Handler that processes Mailgun webhooks.
func (t *Transport) Handler(inbound chan<- core.InboundMessage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if !verifySignature(r.FormValue("timestamp"), r.FormValue("token"), r.FormValue("signature"), t.cfg.SigningKey) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		from := strings.TrimSpace(r.FormValue("sender"))
		if !t.allowed(from) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		text := r.FormValue("stripped-text")
		if text == "" {
			text = r.FormValue("body-plain")
		}
		if len(text) > t.cfg.MaxBytes {
			http.Error(w, "payload too large", http.StatusRequestEntityTooLarge)
			return
		}

		thread := r.FormValue("In-Reply-To")
		if thread == "" {
			thread = r.FormValue("Message-Id")
		}

		inbound <- core.InboundMessage{
			Transport: t.ID(),
			Sender:    from,
			Text:      text,
			ThreadID:  thread,
			Meta: map[string]any{
				"subject":    r.FormValue("subject"),
				"message_id": r.FormValue("Message-Id"),
			},
		}

		w.WriteHeader(http.StatusAccepted)
	})
}

// Send uses Mailgun Messages API.
func (t *Transport) Send(ctx context.Context, msg core.OutboundMessage) error {
	return t.client.Send(ctx, SendRequest{
		From:       "buddy@" + t.cfg.Domain,
		To:         msg.Recipient,
		Subject:    "buddy reply",
		Text:       msg.Text,
		InReplyTo:  msg.ThreadID,
		References: msg.ThreadID,
	})
}

func (t *Transport) allowed(sender string) bool {
	s := strings.ToLower(strings.TrimSpace(sender))
	for _, a := range t.cfg.AllowSenders {
		if strings.ToLower(a) == s {
			return true
		}
	}
	return false
}

// Client wraps minimal Mailgun send.
type Client struct {
	baseURL string
	domain  string
	apiKey  string
	timeout time.Duration
}

func NewClient(baseURL, domain, apiKey string, timeout time.Duration) *Client {
	return &Client{baseURL: baseURL, domain: domain, apiKey: apiKey, timeout: timeout}
}

type SendRequest struct {
	From       string
	To         string
	Subject    string
	Text       string
	InReplyTo  string
	References string
}

func (c *Client) Send(ctx context.Context, req SendRequest) error {
	// Minimal implementation uses Mailgun messages endpoint.
	form := url.Values{}
	form.Set("from", req.From)
	form.Set("to", req.To)
	form.Set("subject", req.Subject)
	form.Set("text", req.Text)
	if req.InReplyTo != "" {
		form.Set("h:In-Reply-To", req.InReplyTo)
	}
	if req.References != "" {
		form.Set("h:References", req.References)
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	endpoint := c.baseURL + "/" + c.domain + "/messages"
	reqHTTP, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	reqHTTP.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqHTTP.SetBasicAuth("api", c.apiKey)

	resp, err := http.DefaultClient.Do(reqHTTP)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("mailgun send failed: %s", resp.Status)
	}
	return nil
}
