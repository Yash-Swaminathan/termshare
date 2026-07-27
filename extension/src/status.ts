export type ShareState = "idle" | "sharing" | "error";

export interface StatusView {
  text: string;
  tooltip: string;
}

// statusView maps a share state to the status bar label/tooltip. Kept pure so
// it can be unit-tested without the VS Code API.
export function statusView(state: ShareState, detail?: string): StatusView {
  switch (state) {
    case "sharing":
      return {
        text: "$(broadcast) termshare: sharing",
        tooltip: "termshare is sharing — click to stop"
      };
    case "error":
      return {
        text: "$(warning) termshare",
        tooltip: detail ? `termshare error: ${detail}` : "termshare error — click to start"
      };
    default:
      return {
        text: "$(terminal) termshare",
        tooltip: "Start a termshare session"
      };
  }
}
