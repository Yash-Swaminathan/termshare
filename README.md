# termshare

Live-share a terminal over WebSockets to a browser. Run one command, get a
shareable link, and let someone on another device watch your shell in real time
(and type, if you let them).

## Demo

![termshare demo](demo/termshare-demo.gif)

<!-- Recording steps and shot list: demo/demo-script.md -->

## Requirements

- Linux or macOS (or Windows via WSL) — the PTY backend is Unix-only
- Go 1.24+
- A trusted network. There is no authentication beyond the host key; this is
  meant for local or trusted-network use.

## Quick start

```bash
go run .
```

On startup termshare creates one session and prints its links:

```text
termshare listening on :8080
session  a1b2c3d4e5f67890
local viewer  http://localhost:8080/s/a1b2c3d4e5f67890
local host    http://localhost:8080/s/a1b2c3d4e5f67890?key=9f8e7d6c5b4a3210
lan viewer    http://192.168.1.42:8080/s/a1b2c3d4e5f67890
lan host      http://192.168.1.42:8080/s/a1b2c3d4e5f67890?key=9f8e7d6c5b4a3210
```

- **Local demo (e.g. recording a video):** open the `local host` URL in one
  browser window and the `local viewer` URL in another on the same machine.
  Type in the host window and watch it mirror live in the viewer.
- **Share with another person:** send them the `lan viewer` URL. They must be on
  the same network (same Wi-Fi / VPN). You drive from a host URL.

The `lan` lines only appear when a non-loopback IPv4 address is detected.

## URL model

| URL | Role | Capabilities |
|-----|------|--------------|
| `/s/{id}` | Viewer | Watch the terminal; read-only unless the host grants write |
| `/s/{id}?key=<hostKey>` | Host | Type, resize the terminal, toggle viewer write access |

- The **session id** is a public room handle — anyone with it can join as a viewer.
- The **host key** is the secret write/admin token. Keep it out of the viewer link.

The host can allow or lock viewer typing at any time with the "Allow viewers to
type" toggle in the UI.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-addr` | `:8080` | HTTP listen address |
| `-host-key` | random | Key that grants host (write) access; random hex if empty |
| `-print-json` | off | Print one JSON line to stdout with `viewer`, `host`, and `id` (for tooling / the VS Code extension) |

## Build

```bash
./scripts/build.sh          # Linux/macOS -> dist/termshare
# or: pwsh ./scripts/build.ps1   # Windows -> dist/termshare.exe
```

Cross-compile (cgo not required):

```bash
GOOS=linux  GOARCH=amd64 go build -o dist/termshare-linux-amd64 .
GOOS=darwin GOARCH=amd64 go build -o dist/termshare-darwin-amd64 .
GOOS=darwin GOARCH=arm64 go build -o dist/termshare-darwin-arm64 .
```

## Limitations

- Unix PTY only (Windows needs WSL or a remote Unix host)
- Reachability depends on the network — LAN sharing needs same Wi-Fi/VPN; no
  built-in tunneling or HTTPS
- No user accounts or OAuth — the host key is the only access control
- One shared shell per process. Run another process for a second session.

## Project layout

```
termshare/
├── main.go           # HTTP server, routing, session URL printing
├── hub.go            # session registry (id -> Session)
├── session.go        # session/room hub: clients, roles, fan-out
├── pty.go            # PTY spawning and resize
├── scripts/          # build helpers for dist/ binaries
├── static/
│   └── index.html    # xterm.js frontend
└── go.mod
```
