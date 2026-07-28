export interface ShareURLs {
  viewer: string;
  host: string;
  id: string;
  lanViewer?: string;
  lanHost?: string;
}

/** Prefer LAN links when present so clipboard/browser work on the same Wi-Fi. */
export function preferredShareLinks(urls: ShareURLs): { viewer: string; host: string } {
  return {
    viewer: urls.lanViewer || urls.viewer,
    host: urls.lanHost || urls.host
  };
}

export function parseShareJSON(line: string): ShareURLs {
  const obj = JSON.parse(line) as Record<string, unknown>;
  const viewer = obj.viewer;
  const host = obj.host;
  const id = obj.id;
  const lanViewer = obj.lanViewer;
  const lanHost = obj.lanHost;

  if (typeof viewer !== "string" || typeof host !== "string" || typeof id !== "string") {
    throw new Error("share JSON missing string fields viewer/host/id");
  }
  if (viewer.includes("key=")) {
    throw new Error("viewer URL must not contain the host key");
  }
  if (!host.includes("key=")) {
    throw new Error("host URL must contain the host key");
  }

  const out: ShareURLs = { viewer, host, id };

  if (lanViewer !== undefined || lanHost !== undefined) {
    if (typeof lanViewer !== "string" || typeof lanHost !== "string") {
      throw new Error("lanViewer and lanHost must both be strings when present");
    }
    if (lanViewer.includes("key=")) {
      throw new Error("lanViewer URL must not contain the host key");
    }
    if (!lanHost.includes("key=")) {
      throw new Error("lanHost URL must contain the host key");
    }
    out.lanViewer = lanViewer;
    out.lanHost = lanHost;
  }

  return out;
}
