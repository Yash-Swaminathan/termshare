import * as vscode from "vscode";
import { ShareSession } from "./session";
import { LaunchMode } from "./launch";
import { preferredShareLinks } from "./parse";
import { detectHostLANIP } from "./lan";
import { startLANBridge } from "./lanbridge";
import { wslPrimaryIP } from "./wslip";
import { ShareState, statusView } from "./status";

let session: ShareSession | undefined;
let statusItem: vscode.StatusBarItem | undefined;
let lanBridge: { close: () => void } | undefined;

export function activate(context: vscode.ExtensionContext): void {
  session = new ShareSession();
  session.onEnded = () => {
    closeLANBridge();
    setStatus("idle");
    vscode.window.showInformationMessage("termshare: share ended (the termshare process exited).");
  };

  statusItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Left, 100);
  statusItem.command = "termshare.toggleShare";
  setStatus("idle");
  statusItem.show();

  context.subscriptions.push(
    statusItem,
    vscode.commands.registerCommand("termshare.startShare", () => startShare(context)),
    vscode.commands.registerCommand("termshare.stopShare", stopShare),
    vscode.commands.registerCommand("termshare.toggleShare", () => toggleShare(context))
  );
}

function setStatus(state: ShareState, detail?: string): void {
  if (!statusItem) {
    return;
  }
  const view = statusView(state, detail);
  statusItem.text = view.text;
  statusItem.tooltip = view.tooltip;
}

function closeLANBridge(): void {
  lanBridge?.close();
  lanBridge = undefined;
}

async function toggleShare(context: vscode.ExtensionContext): Promise<void> {
  if (session?.isRunning()) {
    await stopShare();
  } else {
    await startShare(context);
  }
}

async function startShare(context: vscode.ExtensionContext): Promise<void> {
  if (session?.isRunning()) {
    vscode.window.showInformationMessage("termshare is already sharing.");
    return;
  }

  const config = vscode.workspace.getConfiguration("termshare");

  try {
    const lanIP = detectHostLANIP();
    const addr = config.get<string>("addr", "");
    const port = Number(portFromAddr(addr));

    const urls = await session!.start({
      platform: process.platform,
      arch: process.arch,
      launchMode: config.get<LaunchMode>("launchMode", "bundled"),
      pathSetting: config.get<string>("path", ""),
      binRoot: context.asAbsolutePath("bin"),
      addr,
      hostKey: config.get<string>("hostKey", ""),
      lanIP
    });

    if (process.platform === "win32" && lanIP) {
      try {
        const wslIP = await wslPrimaryIP();
        lanBridge = await startLANBridge({
          lanIP,
          port,
          targetHost: wslIP || "127.0.0.1",
          targetPort: port
        });
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        vscode.window.showWarningMessage(
          `termshare: LAN bridge failed (${message}). Other devices may not reach the LAN link; localhost still works.`
        );
      }
    }

    setStatus("sharing");
    const links = preferredShareLinks(urls);
    await vscode.env.clipboard.writeText(links.viewer);
    await vscode.env.openExternal(vscode.Uri.parse(links.host));
    const kind = urls.lanViewer ? "LAN" : "local";
    vscode.window.showInformationMessage(
      `termshare: sharing started — ${kind} viewer link copied; host opened locally.`
    );
  } catch (err) {
    closeLANBridge();
    const message = err instanceof Error ? err.message : String(err);
    setStatus("error", message);
    vscode.window.showErrorMessage(`termshare: failed to start share. ${message}`);
  }
}

function portFromAddr(addr: string): string {
  const trimmed = addr.trim();
  if (!trimmed) {
    return "8080";
  }
  const idx = trimmed.lastIndexOf(":");
  if (idx >= 0 && trimmed.slice(idx + 1)) {
    return trimmed.slice(idx + 1);
  }
  return "8080";
}

async function stopShare(): Promise<void> {
  if (!session?.isRunning()) {
    vscode.window.showInformationMessage("termshare is not sharing.");
    return;
  }
  closeLANBridge();
  await session.stop();
  setStatus("idle");
  vscode.window.showInformationMessage("termshare: share stopped.");
}

export async function deactivate(): Promise<void> {
  closeLANBridge();
  await session?.stop();
  session = undefined;
}
