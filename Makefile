.PHONY: run build lint test fmt install

CONFIG ?= config.yaml
BIN ?= bin/nostr-codex-runner

run:
	CONFIG=$(CONFIG) go run ./cmd/runner

build:
	go build -o $(BIN) ./cmd/runner

lint:
	go vet ./...

test:
	go test ./...

fmt:
	gofmt -w cmd internal

install:
	go install ./cmd/runner
