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

- PTY support is Linux/macOS only (Windows needs WSL or a remote Unix host)
- No auth — intended for local/trusted network use
- Today every connected client can type into the shared PTY (read-only viewers are Path B, not implemented yet)
- `static/` must stay in the repo — the UI is required to run

## Before Path B / C

Do these first so demos and an extension have a solid base:

1. Commit `static/index.html` if it is still untracked
2. Add a real `README.md` (what it is, how to run, Linux/macOS only, no auth)
3. Fix the product story: either implement host/viewer roles (Path B #1) or stop claiming viewers are read-only
4. Record a short demo (GIF/MP4) once resize + roles exist — better for Twitter/portfolio than more features
5. Decide host OS for demos: PTY will not run natively on Windows; use WSL, macOS, or Linux

Rough effort: Path B ~1–2 days (B-lite: roles + resize ≈ half day). Path C VS Code extension ~1–2 days MVP, ~3–5 days solid.

## Path B — Product upgrades

Make the core tool demo-worthy and match the “live share” story:

1. **Host vs viewer roles** — first client (or token) can write; others read-only unless granted write
2. **Terminal resize** — xterm `onResize` / FitAddon → `pty.Setsize` so layout is not broken
3. **Shareable session URL** — e.g. `/s/{id}` and multi-session support
4. **Shell exit cleanup** — notify clients and tear down cleanly when the shell dies
5. **UI chrome** — viewer count + copy-link button (shows up well in screenshots)

Skip for now unless the project grows: auth/OAuth, public Docker deploy, recording/replay, Windows ConPTY.

## Path C — VS Code extension

Goal: a command like “Start Share” that launches the termshare binary, copies a share URL, and opens the browser (later: status bar, stop share, optional webview with xterm).

- Reuse the Go server as a sidecar process; the extension is mostly TypeScript glue
- Bundle or download platform binaries (darwin/linux; Windows via WSL if needed)
- A thin CLI (`termshare serve`) is useful glue even before the extension ships

### Publishing to the VS Code Marketplace

It is fairly easy — there is **no App Store–style human approval** for a normal first publish.

What you need:

1. Microsoft account + create a **publisher** on the [Visual Studio Marketplace](https://marketplace.visualstudio.com/manage)
2. Auth for `vsce` (PAT today, or Entra/OIDC for CI — Microsoft is moving away from long-lived PATs)
3. Valid `package.json` (`name`, `publisher`, `version`, `engines.vscode`, README, icon rules)
4. `npx @vscode/vsce package` locally, then `vsce publish` (or upload the `.vsix`)

What happens on publish:

- Automated malware + secret scanning; the extension stays private until scans pass
- Usually live quickly if the package is clean — not a multi-week review queue
- Optional later: **Verified Publisher** badge (domain ownership + ~6 months on Marketplace) — separate from being allowed to publish

You can also skip the Marketplace at first:

- Share a `.vsix` (`vsce package` → friends install via “Install from VSIX”)
- For Cursor / VSCodium users, Open VSX is a separate registry if you care about that audience

Practical advice: build and sideload the `.vsix` until Path B basics work; publish when the extension actually starts/stops termshare reliably.
