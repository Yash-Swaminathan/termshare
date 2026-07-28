import { execFile } from "node:child_process";
import { promisify } from "node:util";

const execFileAsync = promisify(execFile);

/** Resolve the primary WSL IPv4 (eth0). */
export async function wslPrimaryIP(): Promise<string | undefined> {
  try {
    const { stdout } = await execFileAsync("wsl", [
      "-e",
      "bash",
      "-lc",
      "hostname -I 2>/dev/null | awk '{print $1}'"
    ]);
    const ip = stdout.trim().split(/\s+/)[0];
    return ip || undefined;
  } catch {
    return undefined;
  }
}
