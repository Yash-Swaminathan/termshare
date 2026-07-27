import * as fs from "node:fs";
import * as path from "node:path";

export type LaunchMode = "bundled" | "path" | "wsl";

export interface SpawnPlan {
  command: string;
  args: string[];
  cwd?: string;
}

export interface ResolveOptions {
  platform: NodeJS.Platform;
  arch: string;
  launchMode: LaunchMode;
  pathSetting: string;
  binRoot: string;
  addr?: string;
  hostKey?: string;
  extraArgs?: string[];
}

// toWSLPath converts a Windows path (C:\a\b) to a WSL mount path (/mnt/c/a/b).
export function toWSLPath(winPath: string): string {
  const m = /^([A-Za-z]):[\\/](.*)$/.exec(winPath);
  if (!m) {
    return winPath;
  }
  const drive = m[1].toLowerCase();
  const rest = m[2].replace(/\\/g, "/");
  return `/mnt/${drive}/${rest}`;
}

// platformDir maps the host to the bundled binary folder name. Windows always
// runs the linux binary under WSL, so the caller passes linux-amd64 directly.
export function platformDir(platform: NodeJS.Platform, arch: string): string {
  const os = platform === "darwin" ? "darwin" : "linux";
  const cpu = arch === "arm64" ? "arm64" : "amd64";
  return `${os}-${cpu}`;
}

function bundledBin(binRoot: string, dir: string): string {
  return path.join(binRoot, dir, "termshare");
}

function requireFile(p: string): void {
  if (!fs.existsSync(p)) {
    throw new Error(
      `termshare binary not found at ${p}. Build the bundled binaries with scripts/build-extension-bins.sh, or set termshare.path / termshare.launchMode.`
    );
  }
}

function runArgs(opts: ResolveOptions): string[] {
  const args = ["-print-json"];
  if (opts.addr && opts.addr.trim() !== "") {
    args.push("-addr", opts.addr.trim());
  }
  if (opts.hostKey && opts.hostKey.trim() !== "") {
    args.push("-host-key", opts.hostKey.trim());
  }
  return [...args, ...(opts.extraArgs ?? [])];
}

export function resolveSpawn(opts: ResolveOptions): SpawnPlan {
  const args = runArgs(opts);
  const onWindows = opts.platform === "win32";
  const setting = opts.pathSetting.trim();

  switch (opts.launchMode) {
    case "path": {
      const bin = setting === "" ? "termshare" : setting;
      if (onWindows) {
        return { command: "wsl", args: ["-e", toWSLPath(bin), ...args] };
      }
      return { command: bin, args };
    }

    case "wsl": {
      if (setting !== "") {
        return { command: "wsl", args: ["-e", toWSLPath(setting), ...args] };
      }
      const bin = bundledBin(opts.binRoot, "linux-amd64");
      requireFile(bin);
      return { command: "wsl", args: ["-e", toWSLPath(bin), ...args] };
    }

    case "bundled":
    default: {
      if (onWindows) {
        const bin = bundledBin(opts.binRoot, "linux-amd64");
        requireFile(bin);
        return { command: "wsl", args: ["-e", toWSLPath(bin), ...args] };
      }
      const bin = bundledBin(opts.binRoot, platformDir(opts.platform, opts.arch));
      requireFile(bin);
      return { command: bin, args };
    }
  }
}
