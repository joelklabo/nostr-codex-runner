#!/usr/bin/env bash
set -euo pipefail

if ! command -v brew >/dev/null 2>&1; then
  echo "Homebrew is required to run this check" >&2
  exit 1
fi

TAP_DIR=$(mktemp -d)
cleanup() {
  rm -rf "$TAP_DIR"
}
trap cleanup EXIT

export HOMEBREW_GITHUB_API_TOKEN=${HOMEBREW_GITHUB_API_TOKEN:-${GH_TOKEN:-}}

echo "Updating brew…"
brew update

echo "Cloning tap to temporary dir: $TAP_DIR"
git clone https://github.com/joelklabo/homebrew-tap "$TAP_DIR"

if brew list --formula | grep -q '^buddy$'; then
  echo "Removing existing buddy formula before retapping"
  brew uninstall buddy
fi

if brew tap | grep -q '^joelklabo/tap$'; then
  echo "Untapping existing joelklabo/tap to avoid remote mismatch"
  brew untap joelklabo/tap
fi

brew tap joelklabo/tap "$TAP_DIR"

brew info joelklabo/tap/buddy
brew install joelklabo/tap/buddy
brew test joelklabo/tap/buddy || {
  echo "brew test reported a failure" >&2
  exit 1
}

buddy version

# Keep a clean workstation
brew uninstall buddy
brew untap joelklabo/tap
