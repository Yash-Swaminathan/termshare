import { test } from "node:test";
import assert from "node:assert/strict";
import { createServer, connect } from "node:net";
import { startLANBridge } from "../lanbridge";

test("LAN bridge forwards TCP to localhost target", async () => {
  const backend = createServer((socket) => {
    socket.write("pong");
    socket.end();
  });
  await new Promise<void>((resolve) => backend.listen(0, "127.0.0.1", resolve));
  const targetPort = (backend.address() as { port: number }).port;

  const probe = createServer();
  await new Promise<void>((resolve) => probe.listen(0, "127.0.0.1", resolve));
  const listenPort = (probe.address() as { port: number }).port;
  await new Promise<void>((resolve) => probe.close(() => resolve()));

  const bridge = await startLANBridge({
    lanIP: "127.0.0.1",
    port: listenPort,
    targetHost: "127.0.0.1",
    targetPort
  });

  const got = await new Promise<string>((resolve, reject) => {
    const sock = connect(listenPort, "127.0.0.1");
    let data = "";
    sock.on("data", (c) => {
      data += c.toString();
    });
    sock.on("end", () => resolve(data));
    sock.on("error", reject);
  });

  assert.equal(got, "pong");
  bridge.close();
  backend.close();
});
