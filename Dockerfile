# Simple Dockerfile for nostr-codex-runner
FROM golang:1.24 as builder
WORKDIR /src
COPY . .
RUN go build -o /out/nostr-codex-runner ./cmd/runner

FROM debian:stable-slim
RUN useradd -m runner
WORKDIR /home/runner
COPY --from=builder /out/nostr-codex-runner /usr/local/bin/nostr-codex-runner
USER runner
ENTRYPOINT ["/usr/local/bin/nostr-codex-runner"]
CMD ["-config", "/home/runner/config.yaml"]
