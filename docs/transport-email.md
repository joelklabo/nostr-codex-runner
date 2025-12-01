# Email Transport Design

Goal: Add email as a transport for buddy. Phase 1 uses Mailgun inbound webhooks (push). Phase 2 adds IMAP/SMTP fallback (polling).

## Why Mailgun first

- Push-based webhooks (no polling), built-in signature verification, retries, good docs.
- Simple config: domain, API key, signing key, route.

## Config (proposed)

```yaml
transports:
  - type: email
    mode: mailgun            # or imap
    id: email
    allow_senders: ["alice@example.com"]
    mailgun:
      domain: mg.example.com
      api_key: $MAILGUN_API_KEY
      signing_key: $MAILGUN_SIGNING_KEY
      base_url: https://api.mailgun.net/v3
      route_prefix: buddy+    # optional; route to POST /email/inbound
    limits:
      max_bytes: 262144
```

IMAP mode (phase 2):

```yaml
    mode: imap
    imap:
      host: imap.example.com
      port: 993
      username: inbox@example.com
      password: $IMAP_PASSWORD
      folder: INBOX
      idle: true
    smtp:
      host: smtp.example.com
      port: 587
      username: inbox@example.com
      password: $SMTP_PASSWORD
```

## Mapping

- Inbound: `from`, `subject` + `text/plain` body → `InboundMessage{Transport:"email", Sender, Text, ThreadID}` where `ThreadID = Message-Id` or `In-Reply-To`.
- Outbound: `Recipient` becomes `to`; set `In-Reply-To` to inbound `Message-Id` to keep threads.
- Strip/ignore HTML; cap size; drop attachments unless enabled later.

## Testing

- Unit: signature verification (Mailgun HMAC), payload mapping, size/allowlist filters.
- Integration (manual): Mailgun sandbox domain + `ngrok http 8080` to hit local handler; send test email and assert reply received.
- IMAP: mock server using `emersion/go-imap/backend` for unit; real mailbox smoke with test creds.

## Health/metrics

- Mailgun: simple `GET /health` returning ok; track counts of accepted/blocked; log signature failures.
- IMAP: NOOP/IDLE keepalive; expose lag since last message; SMTP dial test.

## Cost/ops

- Mailgun inbound typically needs a paid plan after trial; free sandbox for dev. IMAP/SMTP works with any existing mailbox (no extra cost) but requires polling and stored creds.
