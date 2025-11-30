.PHONY: run build lint lint-ci test fmt install verify

CONFIG ?= config.yaml
BIN ?= bin/nostr-codex-runner

run:
	CONFIG=$(CONFIG) go run ./cmd/runner

build:
	go build -o $(BIN) ./cmd/runner

lint:
	go vet ./...

lint-ci:
	@command -v golangci-lint >/dev/null 2>&1 || (echo "golangci-lint not installed; install with 'go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest'"; exit 1)
	golangci-lint run --modules-download-mode=mod --timeout=5m ./...

test:
	go test ./...

fmt:
	gofmt -w cmd internal

install:
	go install ./cmd/runner

verify:
	@gofmt -l cmd internal | tee /tmp/gofmt.out
	@test ! -s /tmp/gofmt.out || (echo "gofmt needed" && cat /tmp/gofmt.out && false)
	go vet ./...
	@command -v staticcheck >/dev/null 2>&1 || go install honnef.co/go/tools/cmd/staticcheck@latest
	staticcheck ./...
	go test -race -covermode=atomic ./...
