export interface ShareURLs {
  viewer: string;
  host: string;
  id: string;
}

export function parseShareJSON(line: string): ShareURLs {
  const obj = JSON.parse(line) as Record<string, unknown>;
  const viewer = obj.viewer;
  const host = obj.host;
  const id = obj.id;

  if (typeof viewer !== "string" || typeof host !== "string" || typeof id !== "string") {
    throw new Error("share JSON missing string fields viewer/host/id");
  }
  if (viewer.includes("key=")) {
    throw new Error("viewer URL must not contain the host key");
  }
  if (!host.includes("key=")) {
    throw new Error("host URL must contain the host key");
  }
  return { viewer, host, id };
}
