# Simple Dockerfile for buddy
FROM golang:1.24.10 AS builder
WORKDIR /src
COPY . .
RUN go build -o /out/buddy ./cmd/runner

FROM debian:stable-slim
RUN useradd -m runner
WORKDIR /home/runner
COPY --from=builder /out/buddy /usr/local/bin/buddy
USER runner
ENTRYPOINT ["/usr/local/bin/buddy"]
CMD ["-config", "/home/runner/config.yaml"]
