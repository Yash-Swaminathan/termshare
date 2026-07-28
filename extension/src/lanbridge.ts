import { createServer, connect } from "node:net";

/**
 * Listen on the Windows LAN interface and tunnel TCP into WSL.
 * Waits for the upstream socket to connect before piping so the
 * WebSocket handshake and early frames are not dropped.
 */
export function startLANBridge(opts: {
  lanIP: string;
  port: number;
  targetHost?: string;
  targetPort?: number;
}): Promise<{ close: () => void }> {
  const targetHost = opts.targetHost ?? "127.0.0.1";
  const targetPort = opts.targetPort ?? opts.port;

  return new Promise((resolve, reject) => {
    const server = createServer((client) => {
      const upstream = connect(targetPort, targetHost);
      const fail = () => {
        client.destroy();
        upstream.destroy();
      };
      client.on("error", fail);
      upstream.on("error", fail);
      upstream.once("connect", () => {
        client.pipe(upstream);
        upstream.pipe(client);
      });
    });

    server.once("error", reject);
    server.listen(opts.port, opts.lanIP, () => {
      resolve({
        close: () => {
          server.close();
        }
      });
    });
  });
}
