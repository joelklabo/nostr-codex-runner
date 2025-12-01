.PHONY: run build lint lint-ci lint-md test fmt install verify brew-check

CONFIG ?= config.yaml
BIN ?= bin/buddy

init-config:
	@test -f $(CONFIG) && echo "$(CONFIG) already exists" || (cp config.example.yaml $(CONFIG) && echo "Wrote $(CONFIG). Edit secrets before running.")

run:
	CONFIG=$(CONFIG) go run ./cmd/runner

build:
	go build -o $(BIN) ./cmd/runner

lint:
	go vet ./...

lint-ci:
	@command -v golangci-lint >/dev/null 2>&1 || (echo "golangci-lint not installed; install with 'go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest'"; exit 1)
	golangci-lint run --modules-download-mode=mod --timeout=5m ./...

lint-md:
	npx -y markdownlint-cli2@0.19.1 "**/*.md" "!**/node_modules" "!.git" "!.beads" "!vendor"

brew-check:
	HOMEBREW_GITHUB_API_TOKEN=$${HOMEBREW_GITHUB_API_TOKEN:-$${GH_TOKEN:-}} scripts/brew-check.sh

test:
	go test ./...

fmt:
	gofmt -w cmd internal

man:
	@command -v go-md2man >/dev/null 2>&1 || (echo "go-md2man not installed; go install github.com/cpuguy83/go-md2man/v2@latest"; exit 1)
	go-md2man -in docs/man/buddy.1.md -out docs/man/buddy.1

man-clean:
	rm -f docs/man/buddy.1

install:
	go install ./cmd/runner

verify:
	@gofmt -l cmd internal | tee /tmp/gofmt.out
	@test ! -s /tmp/gofmt.out || (echo "gofmt needed" && cat /tmp/gofmt.out && false)
	go vet ./...
	@command -v staticcheck >/dev/null 2>&1 || go install honnef.co/go/tools/cmd/staticcheck@latest
	staticcheck ./...
	go test -race -covermode=atomic ./...
