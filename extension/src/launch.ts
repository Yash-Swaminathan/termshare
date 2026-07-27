import * as fs from "node:fs";
import * as path from "node:path";

export interface SpawnPlan {
  command: string;
  args: string[];
  cwd?: string;
}

export interface ResolveOptions {
  platform: NodeJS.Platform;
  pathSetting: string;
  workspaceRoot?: string;
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

export function resolveSpawn(opts: ResolveOptions): SpawnPlan {
  const { platform, pathSetting, workspaceRoot, extraArgs = [] } = opts;
  const setting = pathSetting.trim();

  if (platform === "win32") {
    let bin = setting;
    if (bin === "") {
      const bundled = workspaceRoot
        ? path.join(workspaceRoot, "dist", "termshare-linux-amd64")
        : "";
      bin = bundled && fs.existsSync(bundled) ? toWSLPath(bundled) : "termshare";
    }
    return { command: "wsl", args: ["-e", bin, "-print-json", ...extraArgs], cwd: workspaceRoot };
  }

  const bin = setting === "" ? "termshare" : setting;
  return { command: bin, args: ["-print-json", ...extraArgs], cwd: workspaceRoot };
}
