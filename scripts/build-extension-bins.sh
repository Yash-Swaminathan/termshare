#!/usr/bin/env bash
# Cross-compile termshare into extension/bin/<platform-arch>/termshare so the
# VS Code extension can bundle a self-contained binary per platform.
# cgo is not required (creack/pty is pure Go).
set -euo pipefail
cd "$(dirname "$0")/.."

out="extension/bin"
build() {
  local goos="$1" goarch="$2" dir="$3"
  mkdir -p "$out/$dir"
  GOOS="$goos" GOARCH="$goarch" go build -o "$out/$dir/termshare" .
  echo "built $out/$dir/termshare"
}

build linux  amd64 linux-amd64
build linux  arm64 linux-arm64
build darwin amd64 darwin-amd64
build darwin arm64 darwin-arm64
