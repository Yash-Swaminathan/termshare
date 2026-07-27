import { test } from "node:test";
import assert from "node:assert/strict";
import * as fs from "node:fs";
import * as path from "node:path";
import * as http from "node:http";
import { spawnSync } from "node:child_process";
import { ShareSession } from "../session";

const repoRoot = path.resolve(__dirname, "..", "..", "..");
const linuxBin = path.join(repoRoot, "dist", "termshare-linux-amd64");

function wslAvailable(): boolean {
  if (process.platform !== "win32") {
    return false;
  }
  const r = spawnSync("wsl", ["-e", "true"], { timeout: 10_000 });
  return r.status === 0;
}

const canRun = process.platform === "win32" && fs.existsSync(linuxBin) && wslAvailable();

function get(url: string): Promise<number> {
  return new Promise((resolve, reject) => {
    const req = http.get(url, (res) => {
      res.resume();
      resolve(res.statusCode ?? 0);
    });
    req.on("error", reject);
    req.setTimeout(3_000, () => req.destroy(new Error("http timeout")));
  });
}

async function getWithRetry(url: string, attempts: number): Promise<number> {
  let lastErr: unknown;
  for (let i = 0; i < attempts; i++) {
    try {
      return await get(url);
    } catch (err) {
      lastErr = err;
      await new Promise((r) => setTimeout(r, 300));
    }
  }
  throw lastErr;
}

test(
  "spawns termshare via wsl, serves the viewer page, and stops",
  { skip: canRun ? false : "requires Windows + WSL + dist/termshare-linux-amd64" },
  async () => {
    const session = new ShareSession();
    const urls = await session.start({
      platform: "win32",
      pathSetting: "",
      workspaceRoot: repoRoot,
      extraArgs: ["-addr", ":18790", "-host-key", "itest"]
    });

    try {
      assert.ok(urls.id.length > 0, "expected a session id");
      assert.match(urls.viewer, /:18790\/s\//);
      assert.match(urls.host, /key=itest$/);
      assert.ok(!urls.viewer.includes("key="), "viewer must not leak the key");

      const status = await getWithRetry(urls.viewer, 10);
      assert.equal(status, 200, "viewer page should return 200");
    } finally {
      await session.stop();
    }

    assert.equal(session.isRunning(), false);
    const afterStop = await getWithRetry(urls.viewer, 3).catch(() => "down");
    assert.equal(afterStop, "down", "server should be gone after stop");
  }
);
