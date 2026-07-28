import { test } from "node:test";
import assert from "node:assert/strict";
import * as fs from "node:fs";
import * as os from "node:os";
import * as path from "node:path";
import { resolveSpawn, toWSLPath, platformDir, ResolveOptions } from "../launch";

// makeBinRoot creates a temp bin/ tree with the given platform-arch folders so
// bundled resolution can find a real file.
function makeBinRoot(dirs: string[]): string {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), "termshare-bin-"));
  for (const d of dirs) {
    fs.mkdirSync(path.join(root, d), { recursive: true });
    fs.writeFileSync(path.join(root, d, "termshare"), "");
  }
  return root;
}

function opts(partial: Partial<ResolveOptions>): ResolveOptions {
  return {
    platform: "linux",
    arch: "x64",
    launchMode: "bundled",
    pathSetting: "",
    binRoot: "/nonexistent",
    ...partial
  };
}

test("bundled unix spawns the matching binary directly", () => {
  const binRoot = makeBinRoot(["linux-amd64"]);
  const plan = resolveSpawn(opts({ platform: "linux", arch: "x64", binRoot }));
  assert.equal(plan.command, path.join(binRoot, "linux-amd64", "termshare"));
  assert.deepEqual(plan.args, ["-print-json"]);
});

test("bundled windows runs the linux binary via wsl", () => {
  const binRoot = makeBinRoot(["linux-amd64"]);
  const plan = resolveSpawn(opts({ platform: "win32", arch: "x64", binRoot }));
  assert.equal(plan.command, "wsl");
  assert.equal(plan.args[0], "-e");
  assert.equal(plan.args[1], toWSLPath(path.join(binRoot, "linux-amd64", "termshare")));
  assert.equal(plan.args[plan.args.length - 1], "-print-json");
});

test("bundled throws a clear error when the binary is missing", () => {
  assert.throws(
    () => resolveSpawn(opts({ platform: "linux", binRoot: "/nonexistent" })),
    /termshare binary not found/
  );
});

test("path mode on unix spawns the setting directly", () => {
  const plan = resolveSpawn(opts({ launchMode: "path", pathSetting: "/usr/local/bin/termshare" }));
  assert.equal(plan.command, "/usr/local/bin/termshare");
  assert.deepEqual(plan.args, ["-print-json"]);
});

test("path mode falls back to termshare on PATH", () => {
  const plan = resolveSpawn(opts({ launchMode: "path", pathSetting: "" }));
  assert.equal(plan.command, "termshare");
});

test("path mode on windows wraps in wsl", () => {
  const plan = resolveSpawn(opts({ platform: "win32", launchMode: "path", pathSetting: "termshare" }));
  assert.equal(plan.command, "wsl");
  assert.deepEqual(plan.args, ["-e", "termshare", "-print-json"]);
});

test("wsl mode uses the path setting when set", () => {
  const plan = resolveSpawn(opts({ launchMode: "wsl", pathSetting: "termshare" }));
  assert.equal(plan.command, "wsl");
  assert.deepEqual(plan.args, ["-e", "termshare", "-print-json"]);
});

test("wsl mode falls back to the bundled linux-amd64 binary", () => {
  const binRoot = makeBinRoot(["linux-amd64"]);
  const plan = resolveSpawn(opts({ launchMode: "wsl", binRoot }));
  assert.equal(plan.command, "wsl");
  assert.equal(plan.args[1], toWSLPath(path.join(binRoot, "linux-amd64", "termshare")));
});

test("addr hostKey and lanIP are threaded through as flags", () => {
  const binRoot = makeBinRoot(["linux-amd64"]);
  const plan = resolveSpawn(
    opts({ binRoot, addr: ":9000", hostKey: "abc", lanIP: "192.168.1.5" })
  );
  assert.deepEqual(plan.args, [
    "-print-json",
    "-addr",
    ":9000",
    "-host-key",
    "abc",
    "-lan-ip",
    "192.168.1.5"
  ]);
});

test("platformDir maps os and arch", () => {
  assert.equal(platformDir("darwin", "arm64"), "darwin-arm64");
  assert.equal(platformDir("linux", "x64"), "linux-amd64");
});

test("toWSLPath converts drive paths", () => {
  assert.equal(toWSLPath("C:\\termshare\\dist\\termshare-linux-amd64"), "/mnt/c/termshare/dist/termshare-linux-amd64");
});
