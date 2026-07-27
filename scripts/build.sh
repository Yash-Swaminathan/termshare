#!/usr/bin/env bash
# Build termshare for the current OS into dist/.
# cgo is not required (creack/pty is pure Go).
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p dist
go build -o dist/termshare .
echo "built dist/termshare"

# Cross-compile examples (run manually as needed):
#   GOOS=linux  GOARCH=amd64 go build -o dist/termshare-linux-amd64 .
#   GOOS=darwin GOARCH=amd64 go build -o dist/termshare-darwin-amd64 .
#   GOOS=darwin GOARCH=arm64 go build -o dist/termshare-darwin-arm64 .
#   GOOS=linux  GOARCH=arm64 go build -o dist/termshare-linux-arm64 .
