# termshare

A lightweight terminal live share tool written in Go. Spawn a session and share your terminal with viewers in real time via a browser.

## Stack

- **Go** — WebSocket server + PTY management
- **gorilla/websocket** — WebSocket handling
- **creack/pty** — PTY spawning and I/O
- **xterm.js** — browser-side terminal rendering

## Project Structure

```
termshare/
├── main.go          # entry point, HTTP + WebSocket server
├── session.go       # session/room management
├── pty.go           # PTY spawning and I/O piping
├── static/
│   └── index.html   # xterm.js frontend
└── go.mod
```

## Running Locally

```bash
go run .
# open http://localhost:8080
```

## Notes

- PTY support is Linux/macOS only
- Viewers are read-only by default
- No auth — intended for local/trusted network use