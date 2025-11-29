.PHONY: run build lint

run:
	CONFIG=config.yaml go run ./cmd/runner

build:
	go build -o bin/nostr-codex-runner ./cmd/runner

lint:
	go vet ./...
