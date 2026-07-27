# Build termshare for the current OS into dist/.
# cgo is not required (creack/pty is pure Go).
$ErrorActionPreference = "Stop"
Set-Location (Split-Path $PSScriptRoot -Parent)
New-Item -ItemType Directory -Force -Path dist | Out-Null
go build -o dist/termshare.exe .
Write-Host "built dist/termshare.exe"

# Cross-compile examples (run manually as needed):
#   $env:GOOS="linux";  $env:GOARCH="amd64"; go build -o dist/termshare-linux-amd64 .
#   $env:GOOS="darwin"; $env:GOARCH="amd64"; go build -o dist/termshare-darwin-amd64 .
#   $env:GOOS="darwin"; $env:GOARCH="arm64"; go build -o dist/termshare-darwin-arm64 .
#   $env:GOOS="linux";  $env:GOARCH="arm64"; go build -o dist/termshare-linux-arm64 .
#   Remove-Item Env:GOOS, Env:GOARCH
