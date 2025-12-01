# Run as a systemd service (Linux)

This recipe shows how to keep the runner alive on Linux using systemd. It expects a per-user config at `~/.config/nostr-codex-runner/config.yaml` and the binary at `~/.local/bin/nostr-codex-runner`.

## 1) Install the binary and config

```bash
go install github.com/joelklabo/nostr-codex-runner/cmd/runner@latest
mkdir -p ~/.config/nostr-codex-runner
cp config.example.yaml ~/.config/nostr-codex-runner/config.yaml
# edit config.yaml with your secrets and plugin choices
```

## 2) Install the unit

```bash
sudo cp scripts/systemd/nostr-codex-runner.service /etc/systemd/system/nostr-codex-runner@${USER}.service
sudo systemctl daemon-reload
sudo systemctl enable --now nostr-codex-runner@${USER}.service
```

The unit uses:

- `NCR_CONFIG=%h/.config/nostr-codex-runner/config.yaml`
- `ExecStart=%h/.local/bin/nostr-codex-runner run -config ${NCR_CONFIG}`

Adjust paths if you installed elsewhere.

## 3) Check status and logs

```bash
systemctl status nostr-codex-runner@${USER}.service
journalctl -u nostr-codex-runner@${USER}.service -f
```

## 4) Upgrade

```bash
go install github.com/joelklabo/nostr-codex-runner/cmd/runner@latest
sudo systemctl restart nostr-codex-runner@${USER}.service
```

If you change config, reload/restart is enough; no daemon-reload needed unless the unit file changes.
