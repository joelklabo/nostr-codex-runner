# Run with Docker

You can build and run the runner in a container. This is useful for testing or keeping host dependencies minimal.

## Build the image

```bash
docker build -t nostr-codex-runner:local .
```

## Prepare config and state dirs

```bash
mkdir -p $PWD/.data
cp config.example.yaml $PWD/config.yaml
# edit config.yaml with your secrets; keep state under .data/state.db
```

## Run

```bash
docker run --rm \
  -v $PWD/config.yaml:/app/config.yaml:ro \
  -v $PWD/.data:/app/data \
  -e NCR_CONFIG=/app/config.yaml \
  nostr-codex-runner:local \
  ./nostr-codex-runner run -config /app/config.yaml
```

Notes:

- The provided Dockerfile builds a static binary in the image.
- Mount a writable volume for BoltDB state (default `storage.path`).
- Expose any transport ports you need (e.g., WhatsApp webhook): `-p 8083:8083`.
- For health checks, run with `-health-listen 0.0.0.0:8081` and expose that port too.
