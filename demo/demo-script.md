# termshare demo — shot list

A repeatable ~20-40s demo for Twitter/portfolio. Verified working on WSL Ubuntu.

## Setup (host machine = WSL Ubuntu)

1. In WSL, from the repo:

   ```bash
   go run . -host-key demo
   ```

2. Copy the two `local` URLs it prints:
   - `local host   http://localhost:8080/s/<id>?key=demo`
   - `local viewer http://localhost:8080/s/<id>`

   (WSL2 forwards `localhost`, so open these in a normal Windows browser.)

## Window layout for capture

- Left: WSL terminal running the server (shows the printed URLs — nice context).
- Top-right: browser at the **host** URL.
- Bottom-right: browser at the **viewer** URL.

Record with OBS / ScreenToGif / Loom. Target 20-40s, then export `termshare-demo.gif`
(for the README embed) and/or `.mp4` (for Twitter/X). Drop the file in this `demo/` folder.

## Beats (what to do on camera)

1. Show the server URLs in the terminal (1-2s).
2. In the **host** window, type a visible command:

   ```bash
   ls -la && echo "live from termshare"
   ```

   Point out it appears in the **viewer** window in real time.
3. Type something fun for scale:

   ```bash
   neofetch
   ```

   (or `htop`, then `q` — anything with motion reads well on video.)
4. In the **host** window, click **Allow viewers to type**. Switch to the
   **viewer** window and type a command there to show write access was granted.
5. Click **Lock viewers (read-only)** in the host to show control is server-side.
6. End on the repo URL / one-liner: "live-share a PTY over WebSockets in ~250 LOC of Go".

## Caption / copy ideas

- Hook: "Built a tiny tty-share clone in Go — share your terminal to a browser."
- Bullets: WebSocket fan-out - PTY - xterm.js - host/viewer roles + shareable `/s/{id}` links.

## Notes

- The `lan` URLs printed under WSL point at the WSL virtual NIC (e.g. `10.255.x.x`),
  which is NAT'd — sharing to a separate physical device from WSL needs port
  forwarding. For a real cross-device share, run termshare on a native Linux/macOS
  host. For the demo, the `local` URLs on one machine are all you need.
