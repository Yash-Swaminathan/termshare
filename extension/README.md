# termshare VS Code extension

Start a [termshare](https://github.com/Yash-Swaminathan/termshare) session from
VS Code: it launches the termshare server, copies the read-only viewer link to
your clipboard, and opens the host link in your browser. A status bar item shows
whether you are sharing and toggles start/stop.

## Commands

- `termshare: Start Share` - launch the server, copy the viewer URL, open the host URL
- `termshare: Stop Share` - stop the server
- `termshare: Toggle Share` - start or stop (also bound to the status bar item)

## Settings

| Setting | Default | Purpose |
|---------|---------|---------|
| `termshare.launchMode` | `bundled` | How to launch: `bundled` (binary shipped with the extension), `path` (`termshare.path` / `termshare` on `PATH`), or `wsl` (force WSL). |
| `termshare.path` | `""` | Binary path for `path` mode, or an override for `wsl`. |
| `termshare.addr` | `""` | Listen address passed as `-addr` (e.g. `:8080`). |
| `termshare.hostKey` | `""` | Fixed host key passed as `-host-key`; empty means random per session. |

## Requirements

The termshare PTY backend is Unix-only, so on Windows the extension runs the
binary through WSL. In `bundled` mode it ships a self-contained binary per
platform (the UI is embedded, so it no longer needs a sibling `static/` folder):

- Windows: the bundled `bin/linux-amd64/termshare`, run via `wsl`.
- macOS/Linux: the bundled `bin/<platform>-<arch>/termshare`.

Use `path` mode with `termshare.path` to point at your own binary instead.

## Build the bundled binaries

The per-platform binaries are gitignored; build them before packaging:

```bash
scripts/build-extension-bins.sh   # from the repo root; needs Go + bash (WSL on Windows)
```

This cross-compiles into `extension/bin/{linux-amd64,linux-arm64,darwin-amd64,darwin-arm64}/termshare`.

## Develop

```bash
cd extension
npm install
npm run compile   # type-check + bundle to dist/extension.js
npm test          # unit tests
npm run package   # build a .vsix (vsce)
```

Press F5 in VS Code to launch an Extension Development Host and try the commands.
Install the packaged `.vsix` via "Extensions: Install from VSIX".
