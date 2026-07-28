import { createServer, connect } from "node:net";

/**
 * Listen on the Windows LAN interface and tunnel TCP to WSL's localhost relay.
 * Avoids needing elevated netsh portproxy: WSL already publishes the server on
 * 127.0.0.1:<port> via localhost forwarding; other devices need the LAN IP.
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
      client.pipe(upstream);
      upstream.pipe(client);
      const fail = () => {
        client.destroy();
        upstream.destroy();
      };
      client.on("error", fail);
      upstream.on("error", fail);
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
