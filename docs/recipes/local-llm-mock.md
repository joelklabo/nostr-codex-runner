# Local LLM with Mock Transport (offline)

Goal: run buddy fully offline by combining the `local-llm` preset with the mock transport. This avoids nostr keys/relays and is ideal for air‑gapped testing.

## Steps

1. Copy the preset override to your user presets directory:

   ```bash
   mkdir -p ~/.config/buddy/presets
   cat > ~/.config/buddy/presets/local-llm.yaml <<'EOF'
   transports:
     - type: mock
       id: mock
   agent:
     type: http
     config:
       base_url: http://127.0.0.1:11434   # replace with your local endpoint
       model: local-model
       timeout_seconds: 120
   actions: []
   runner:
     private_key: "mock"
     allowed_pubkeys: ["mock"]
     session_timeout_minutes: 240
   storage:
     path: ~/.buddy/state.db
   EOF

   ```

1. Start buddy with the preset name:

   ```bash
   buddy run local-llm
   ```

1. Inject a message through the mock transport (example using tests/scripts) or wire your own harness; nostr DMs are not used in mock mode.

## Notes

- This override is picked up ahead of the embedded preset; remove the file to go back to nostr transport.

- Keep the mock keys (`mock`) for offline use only; they are placeholders to satisfy validation.

- Adjust `base_url`/`model` to your local LLM server; add `actions` if you want shell/fs.
