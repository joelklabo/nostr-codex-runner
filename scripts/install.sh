#!/usr/bin/env bash
set -euo pipefail

REPO="joelklabo/buddy"
BINARY_NAME="buddy"
ALIAS_NAME="nostr-buddy"
INSTALL_DIR=${INSTALL_DIR:-"$HOME/.local/bin"}
VERSION=${VERSION:-"latest"}
CONFIG_DIR=${CONFIG_DIR:-"$(pwd)"}
INSTALL_ALIAS=${INSTALL_ALIAS:-"true"} # set to false to skip alias/symlink

log() { echo "[installer] $*"; }
err() { echo "[installer] error: $*" >&2; }

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || { err "Missing required command: $1"; exit 1; }
}

pick_asset() {
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  arch=$(uname -m)
  case "$arch" in
    x86_64|amd64) arch=amd64;;
    aarch64|arm64) arch=arm64;;
    *) err "Unsupported arch: $arch"; exit 1;;
  esac
  echo "${BINARY_NAME}-${os}-${arch}"
}

fetch_latest_tag() {
  need_cmd curl
  curl -sSfL "https://api.github.com/repos/${REPO}/releases/latest" | \
    grep -m1 '"tag_name"' | sed 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/'
}

download_asset() {
  need_cmd curl
  need_cmd tar
  asset=$(pick_asset)
  url="https://github.com/${REPO}/releases/download/${1}/${asset}"
  tmp=$(mktemp)
  log "Downloading $url"
  if ! curl -fL "$url" -o "$tmp"; then
    err "Download failed"; rm -f "$tmp"; return 1
  fi
  chmod +x "$tmp"
  mkdir -p "$INSTALL_DIR"
  mv "$tmp" "$INSTALL_DIR/$BINARY_NAME"
  log "Installed to $INSTALL_DIR/$BINARY_NAME"
  if [ "$INSTALL_ALIAS" = "true" ]; then
    ln -sf "$INSTALL_DIR/$BINARY_NAME" "$INSTALL_DIR/$ALIAS_NAME"
    log "Alias created: $INSTALL_DIR/$ALIAS_NAME"
  fi
}

fallback_go_install() {
  need_cmd go
  log "Falling back to 'go install'"
  go install github.com/${REPO}/cmd/runner@${1}
  log "Binary on PATH via GOPATH/bin (ensure it's in PATH)"
}

copy_config() {
  example="config.example.yaml"
  target="$CONFIG_DIR/config.yaml"
  if [ -f "$target" ]; then
    log "config.yaml already exists; leaving it"
    return
  fi
  if [ -f "$example" ]; then
    cp "$example" "$target"
    log "Wrote $target (edit runner.private_key and allowed_pubkeys)"
  else
    log "config.example.yaml not found; skipped config copy"
  fi
}

main() {
  if [ "$VERSION" = "latest" ]; then
    VERSION=$(fetch_latest_tag)
    log "Latest tag: $VERSION"
  fi
  if ! download_asset "$VERSION"; then
    fallback_go_install "$VERSION"
  fi
  copy_config
  log "Done. Add $INSTALL_DIR to PATH if not already."
}

main "$@"
