#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

go run ./cmd/runner --config "${CONFIG:-config.yaml}"
