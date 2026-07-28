# termshare

Live-share your terminal from VS Code.

Start Share launches the termshare server, copies the viewer link to your
clipboard, and opens the host link in your browser. The status bar shows when
you are sharing and toggles start/stop.

## Commands

- **termshare: Start Share**
- **termshare: Stop Share**
- **termshare: Toggle Share** (status bar)

## Requirements

termshare needs a Unix PTY (Linux, macOS, or Windows via WSL). Bundled mode
ships platform binaries; on Windows the Linux binary runs through WSL.

## Settings

| Setting | Default | Purpose |
|---------|---------|---------|
| `termshare.launchMode` | `bundled` | `bundled`, `path`, or `wsl` |
| `termshare.path` | `""` | Binary path for `path` / override for `wsl` |
| `termshare.addr` | `""` | Listen address (`-addr`), empty = default |
| `termshare.hostKey` | `""` | Fixed host key (`-host-key`), empty = random |

## More

Source and CLI docs: [github.com/Yash-Swaminathan/termshare](https://github.com/Yash-Swaminathan/termshare)
