import * as os from "node:os";

/** Score LAN candidates: prefer home Wi-Fi (192.168), then other RFC1918. */
export function lanIPScore(ip: string): number {
  const m = /^(\d+)\.(\d+)\.(\d+)\.(\d+)$/.exec(ip);
  if (!m) {
    return 0;
  }
  const a = Number(m[1]);
  const b = Number(m[2]);
  const d = Number(m[4]);
  let score = 5;
  if (a === 192 && b === 168) {
    score = 30;
  } else if (a === 10) {
    score = 20;
  } else if (a === 172 && b >= 16 && b <= 31) {
    score = 10;
  }
  // VirtualBox host-only is rarely the path phones use.
  if (ip.startsWith("192.168.56.")) {
    score -= 20;
  }
  // .1 is often the router/gateway, not this machine.
  if (d === 1) {
    score -= 8;
  }
  return score;
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
    if (!addrs || /loopback|vethernet|hyper-v|docker|wsl|virtualbox|vmware|bluetooth/i.test(name)) {
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
      if (/wi-?fi|wlan|wireless/i.test(name)) {
        score += 15;
      } else if (/ethernet|en0|eth0/i.test(name)) {
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
