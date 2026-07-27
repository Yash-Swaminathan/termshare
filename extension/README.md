# termshare VS Code extension

Start a [termshare](https://github.com/Yash-Swaminathan/termshare) session from
VS Code: it launches the termshare server, copies the read-only viewer link to
your clipboard, and opens the host link in your browser.

## Commands

- `termshare: Start Share` - launch the server, copy the viewer URL, open the host URL
- `termshare: Stop Share` - stop the server

## Requirements

The termshare PTY backend is Unix-only. On Windows the extension runs the binary
through WSL. Point it at a binary with the `termshare.path` setting, or leave it
empty to auto-resolve:

- Windows: `dist/termshare-linux-amd64` in the workspace (run via `wsl`) if
  present, otherwise `termshare` on the WSL `PATH`.
- macOS/Linux: `termshare` on `PATH`.

The termshare binary serves its `static/` UI relative to its working directory,
so it is launched from the workspace root.

## Develop

```bash
cd extension
npm install
npm run compile   # type-check + bundle to dist/extension.js
npm test          # unit tests + WSL integration test (skips without WSL)
```

Press F5 in VS Code to launch an Extension Development Host and try the commands.
