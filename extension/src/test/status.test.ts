import { test } from "node:test";
import assert from "node:assert/strict";
import { statusView } from "../status";

test("idle shows the terminal icon", () => {
  const v = statusView("idle");
  assert.match(v.text, /termshare/);
  assert.match(v.text, /\$\(terminal\)/);
});

test("sharing shows the broadcast icon", () => {
  const v = statusView("sharing");
  assert.match(v.text, /sharing/);
  assert.match(v.text, /\$\(broadcast\)/);
});

test("error includes the detail in the tooltip", () => {
  const v = statusView("error", "port in use");
  assert.match(v.text, /\$\(warning\)/);
  assert.match(v.tooltip, /port in use/);
});
