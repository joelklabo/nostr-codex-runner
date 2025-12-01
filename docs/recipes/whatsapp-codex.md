# Recipe: WhatsApp (Twilio) transport + Codex CLI agent + basic actions

This flow listens on a Twilio WhatsApp webhook, forwards messages to the agent, and replies via Twilio Messages API.

## Prerequisites

- Twilio account with a WhatsApp-enabled number.

- `ACCOUNT_SID`, `AUTH_TOKEN`, and `FROM` number (e.g., `whatsapp:+15550001234`).

- Publicly reachable webhook URL (ngrok/Cloudflare Tunnel/etc.).

- Go 1.22+ or release binary of this runner.

- Codex CLI installed and on `PATH`.

## Configure

Add a transport entry (or copy to a dedicated config file):

```yaml
transports:
  - type: whatsapp
    id: whatsapp
    config:
      account_sid: "ACxxxxxxxx"
      auth_token: "your_twilio_auth_token"
      from_number: "whatsapp:+15550001234"
      listen: ":8083"
      path: "/twilio/webhook"
      allowed_numbers: ["15555550100"]   # optional allowlist (E.164 without +)

agent:
  type: codexcli
  config:
    binary: codex
    working_dir: .
    timeout_seconds: 900

actions:
  - type: shell
    name: shell
    workdir: .
    timeout_seconds: 30
    max_output: 4000
  - type: readfile
    roots: ["."]
    max_bytes: 65536
  - type: writefile
    roots: ["."]
    allow_write: true
    max_bytes: 65536

runner:
  allowed_pubkeys: []   # not used by WhatsApp; keep empty
  max_reply_chars: 4000
  initial_prompt: "You are an agent responding to WhatsApp users. Be concise and safe."

storage:
  path: ./state.db

logging:
  level: info
```

## Wire up Twilio webhook

1. Run the runner: `make run` (or `./bin/buddy -config config.yaml`).

2. Expose port `8083` publicly (e.g., `ngrok http 8083`).

3. In Twilio Console, set the WhatsApp sandbox/number webhook URL to `https://<public-host>/twilio/webhook`.

4. Send a WhatsApp message to your Twilio number from an allowed phone; the runner should reply via the agent.

## Notes

- Signature verification uses Twilio’s `X-Twilio-Signature` with your auth token.

- `base_url` in `config` can point to a mock server for testing.

- The flow is fully composable: swap `agent.type` or `actions` without touching the transport.
