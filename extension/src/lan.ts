import * as os from "node:os";

/** Score LAN candidates: prefer home Wi-Fi (192.168), then other RFC1918. */
export function lanIPScore(ip: string): number {
  const m = /^(\d+)\.(\d+)\.(\d+)\.(\d+)$/.exec(ip);
  if (!m) {
    return 0;
  }
  const a = Number(m[1]);
  const b = Number(m[2]);
  if (a === 192 && b === 168) {
    return 30;
  }
  if (a === 10) {
    return 20;
  }
  if (a === 172 && b >= 16 && b <= 31) {
    return 10;
  }
  return 5;
}

/**
 * Pick a host LAN IPv4 for share links (Windows Wi-Fi / Ethernet).
 * Skips internal/loopback and link-local addresses.
 */
export function detectHostLANIP(
  interfaces: NodeJS.Dict<os.NetworkInterfaceInfo[]> = os.networkInterfaces()
): string | undefined {
  let best: string | undefined;
  let bestScore = 0;
  for (const [name, addrs] of Object.entries(interfaces)) {
    if (!addrs || /loopback|vethernet|hyper-v|docker|wsl/i.test(name)) {
      continue;
    }
    for (const addr of addrs) {
      if (addr.family !== "IPv4" && (addr.family as unknown) !== 4) {
        continue;
      }
      if (addr.internal) {
        continue;
      }
      if (addr.address.startsWith("169.254.")) {
        continue;
      }
      let score = lanIPScore(addr.address);
      if (/wi-?fi|wlan|wireless|ethernet|en0|eth0/i.test(name)) {
        score += 5;
      }
      if (score > bestScore) {
        bestScore = score;
        best = addr.address;
      }
    }
  }
  return best;
}
