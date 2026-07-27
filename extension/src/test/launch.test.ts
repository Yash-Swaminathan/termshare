import { test } from "node:test";
import assert from "node:assert/strict";
import { resolveSpawn, toWSLPath } from "../launch";

test("unix spawns the binary directly", () => {
  const plan = resolveSpawn({ platform: "linux", pathSetting: "" });
  assert.equal(plan.command, "termshare");
  assert.deepEqual(plan.args, ["-print-json"]);
});

test("unix honors the path setting", () => {
  const plan = resolveSpawn({ platform: "darwin", pathSetting: "/usr/local/bin/termshare" });
  assert.equal(plan.command, "/usr/local/bin/termshare");
  assert.deepEqual(plan.args, ["-print-json"]);
});

test("windows wraps the command in wsl", () => {
  const plan = resolveSpawn({ platform: "win32", pathSetting: "" });
  assert.equal(plan.command, "wsl");
  assert.equal(plan.args[0], "-e");
  assert.equal(plan.args[plan.args.length - 1], "-print-json");
});

test("windows path setting wins over auto-resolve", () => {
  const plan = resolveSpawn({ platform: "win32", pathSetting: "termshare-custom" });
  assert.deepEqual(plan.args, ["-e", "termshare-custom", "-print-json"]);
});

test("toWSLPath converts drive paths", () => {
  assert.equal(toWSLPath("C:\\termshare\\dist\\termshare-linux-amd64"), "/mnt/c/termshare/dist/termshare-linux-amd64");
});
