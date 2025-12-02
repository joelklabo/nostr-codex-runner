package imap

import (
	"context"
	"fmt"
	"io"
	"net/smtp"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	imapclient "github.com/emersion/go-imap/client"
	"github.com/emersion/go-message"
	"github.com/joelklabo/buddy/internal/core"
)

// Transport implements a polling IMAP receive + SMTP send.
type Transport struct {
	cfg Config
}

func New(cfg Config) (*Transport, error) {
	cfg.Defaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Transport{cfg: cfg}, nil
}

func (t *Transport) ID() string { return t.cfg.ID }

func (t *Transport) Start(ctx context.Context, inbound chan<- core.InboundMessage) error {
	for {
		if err := t.pollOnce(ctx, inbound); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			// backoff on errors
			select {
			case <-time.After(5 * time.Second):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		select {
		case <-time.After(30 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (t *Transport) pollOnce(ctx context.Context, inbound chan<- core.InboundMessage) error {
	c, err := imapclient.DialTLS(fmt.Sprintf("%s:%d", t.cfg.Host, t.cfg.Port), nil)
	if err != nil {
		return err
	}
	defer c.Logout()

	if err := c.Login(t.cfg.Username, t.cfg.Password); err != nil {
		return err
	}

	if _, err := c.Select(t.cfg.Folder, false); err != nil {
		return err
	}

	seqset := new(imap.SeqSet)
	seqset.AddRange(1, 0)
	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchBody}, messages)
	}()

	for msg := range messages {
		if msg.Envelope == nil {
			continue
		}
		from := ""
		if len(msg.Envelope.From) > 0 {
			from = msg.Envelope.From[0].Address()
		}
		var body strings.Builder
		section := &imap.BodySectionName{}
		if r := msg.GetBody(section); r != nil {
			m, _ := message.Read(r)
			if m != nil {
				b, _ := io.ReadAll(m.Body)
				body.Write(b)
			}
		}
		inbound <- core.InboundMessage{
			Transport: t.ID(),
			Sender:    from,
			Text:      body.String(),
			ThreadID:  msg.Envelope.InReplyTo,
			Meta: map[string]any{
				"subject":    msg.Envelope.Subject,
				"message_id": msg.Envelope.MessageId,
			},
		}
	}
	return <-done
}

func (t *Transport) Send(ctx context.Context, msg core.OutboundMessage) error {
	auth := smtp.PlainAuth("", t.cfg.Username, t.cfg.Password, t.cfg.SMTPHost)
	to := []string{msg.Recipient}
	data := fmt.Sprintf("Subject: buddy reply\r\nIn-Reply-To: %s\r\n\r\n%s", msg.ThreadID, msg.Text)
	return smtp.SendMail(fmt.Sprintf("%s:%d", t.cfg.SMTPHost, t.cfg.SMTPPort), auth, t.cfg.Username, to, []byte(data))
}
