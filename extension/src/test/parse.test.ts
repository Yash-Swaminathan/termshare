import { test } from "node:test";
import assert from "node:assert/strict";
import { parseShareJSON, preferredShareLinks } from "../parse";

test("parses a valid Phase 0 line", () => {
  const line = JSON.stringify({
    viewer: "http://localhost:8080/s/abc",
    host: "http://localhost:8080/s/abc?key=secret",
    id: "abc"
  });
  const got = parseShareJSON(line);
  assert.equal(got.viewer, "http://localhost:8080/s/abc");
  assert.equal(got.host, "http://localhost:8080/s/abc?key=secret");
  assert.equal(got.id, "abc");
  assert.equal(got.lanViewer, undefined);
});

test("parses optional LAN fields", () => {
  const line = JSON.stringify({
    viewer: "http://localhost:8080/s/abc",
    host: "http://localhost:8080/s/abc?key=secret",
    lanViewer: "http://192.168.1.10:8080/s/abc",
    lanHost: "http://192.168.1.10:8080/s/abc?key=secret",
    id: "abc"
  });
  const got = parseShareJSON(line);
  assert.equal(got.lanViewer, "http://192.168.1.10:8080/s/abc");
  assert.equal(got.lanHost, "http://192.168.1.10:8080/s/abc?key=secret");
});

test("preferredShareLinks uses LAN viewer but keeps local host", () => {
  const links = preferredShareLinks({
    viewer: "http://localhost:8080/s/abc",
    host: "http://localhost:8080/s/abc?key=secret",
    lanViewer: "http://192.168.1.10:8080/s/abc",
    lanHost: "http://192.168.1.10:8080/s/abc?key=secret",
    id: "abc"
  });
  assert.equal(links.viewer, "http://192.168.1.10:8080/s/abc");
  assert.equal(links.host, "http://localhost:8080/s/abc?key=secret");
});

test("preferredShareLinks falls back to local", () => {
  const links = preferredShareLinks({
    viewer: "http://localhost:8080/s/abc",
    host: "http://localhost:8080/s/abc?key=secret",
    id: "abc"
  });
  assert.equal(links.viewer, "http://localhost:8080/s/abc");
  assert.equal(links.host, "http://localhost:8080/s/abc?key=secret");
});

test("rejects missing fields", () => {
  assert.throws(() => parseShareJSON(JSON.stringify({ viewer: "x", id: "y" })));
});

test("rejects non-string fields", () => {
  const line = JSON.stringify({ viewer: "x", host: "y?key=z", id: 5 });
  assert.throws(() => parseShareJSON(line));
});

test("rejects viewer that leaks the key", () => {
  const line = JSON.stringify({
    viewer: "http://localhost:8080/s/abc?key=secret",
    host: "http://localhost:8080/s/abc?key=secret",
    id: "abc"
  });
  assert.throws(() => parseShareJSON(line), /viewer/);
});

test("rejects host without a key", () => {
  const line = JSON.stringify({
    viewer: "http://localhost:8080/s/abc",
    host: "http://localhost:8080/s/abc",
    id: "abc"
  });
  assert.throws(() => parseShareJSON(line), /host/);
});

test("rejects non-JSON", () => {
  assert.throws(() => parseShareJSON("not json"));
});
