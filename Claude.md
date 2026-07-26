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
├── main.go          # entry point, HTTP + WebSocket server, session URL printing
├── hub.go           # session registry (id -> Session)
├── session.go       # session/room management
├── pty.go           # PTY spawning and I/O piping
├── static/
│   └── index.html   # xterm.js frontend
└── go.mod
```

## Running Locally

```bash
go run .
# startup prints a session id plus links, e.g.:
#   local viewer  http://localhost:8080/s/<id>
#   local host    http://localhost:8080/s/<id>?key=<hostKey>
#   lan   viewer  http://<lan-ip>:8080/s/<id>          (shown when a LAN IP is found)
#   lan   host    http://<lan-ip>:8080/s/<id>?key=<hostKey>
```

Sessions live at `/s/{id}`. The `?key=<hostKey>` grants host (write) access;
without it you are a read-only viewer. For a local demo, open the `local host`
link in one window and the `local viewer` link in another. To share, send the
`lan viewer` link to someone on the same network.

The host key is random each run; override with `go run . -host-key <key>`.

## Notes

- PTY support is Linux/macOS only (Windows needs WSL or a remote Unix host)
- Sessions are keyed by a public random id and resolved through the `Hub`
  (`hub.go`); one process creates exactly one session at boot. Unknown ids 404.
- Roles are enforced server-side: only a client that connected to `/s/{id}/ws`
  with the correct `?key=<hostKey>` is the host and may type. Everyone else is a
  read-only viewer.
- The host can grant/revoke write for all viewers at once (the "Allow viewers to
  type" toggle sends `{"type":"set_acl","viewersWrite":bool}`); server updates
  each viewer's `canWrite` and pushes a `role` message back.
- No user auth beyond the host key — intended for local/trusted network use
- `static/` must stay in the repo — the UI is required to run

## Before Path B / C

Status of the pre-demo checklist:

1. [x] Commit `static/index.html` — tracked and committed
2. [x] Real `README.md` — added (what it is, how to run, Unix-only, no auth)
3. [x] Product story fixed — host/viewer roles + write ACL implemented (Path B #1)
4. [ ] Record a short demo (GIF/MP4) — shot list ready in `demo/demo-script.md`;
   README embeds `demo/termshare-demo.gif` once the file is dropped in. This is the
   only open pre-demo item (manual screen capture).
5. [x] Host OS for demos: **WSL Ubuntu** (native Windows is unsupported — `pty.Start`
   returns `unsupported`). Verified end-to-end on WSL: boot, `/s/{id}` 200, bad id
   404, and host-typed output reaching a viewer over WebSocket.

Rough effort: Path B ~1–2 days (B-lite: roles + resize ≈ half day). Path C VS Code extension ~1–2 days MVP, ~3–5 days solid.

## Path B — Product upgrades

Make the core tool demo-worthy and match the “live share” story:

1. **Host vs viewer roles** — first client (or token) can write; others read-only unless granted write
2. **Terminal resize** — xterm `onResize` / FitAddon → `pty.Setsize` so layout is not broken
3. **Shareable session URL** — e.g. `/s/{id}` and multi-session support
4. **Shell exit cleanup** (done) — notify clients and tear down cleanly when the shell dies
5. **UI chrome** (done) — viewer count + copy-link button (shows up well in screenshots)

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
