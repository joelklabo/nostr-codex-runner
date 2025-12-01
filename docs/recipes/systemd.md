# Run as a systemd service (Linux)

This recipe shows how to keep the runner alive on Linux using systemd. It expects a per-user config at `~/.config/buddy/config.yaml` and the binary at `~/.local/bin/buddy` (alias `nostr-buddy` if desired).

## 1) Install the binary and config

```bash
go install github.com/joelklabo/buddy/cmd/runner@latest
mkdir -p ~/.config/buddy
cp config.example.yaml ~/.config/buddy/config.yaml
# edit config.yaml with your secrets and plugin choices
```

## 2) Install the unit

```bash
sudo cp scripts/systemd/buddy.service /etc/systemd/system/nostr-buddy@${USER}.service
sudo systemctl daemon-reload
sudo systemctl enable --now nostr-buddy@${USER}.service
```

The unit uses:

- `BUDDY_CONFIG=%h/.config/buddy/config.yaml`
- `ExecStart=%h/.local/bin/buddy run -config ${BUDDY_CONFIG}`

Adjust paths if you installed elsewhere.

## 3) Check status and logs

```bash
systemctl status nostr-buddy@${USER}.service
journalctl -u nostr-buddy@${USER}.service -f
```

## 4) Upgrade

```bash
go install github.com/joelklabo/buddy/cmd/runner@latest
sudo systemctl restart nostr-buddy@${USER}.service
```

If you change config, reload/restart is enough; no daemon-reload needed unless the unit file changes.
