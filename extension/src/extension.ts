import * as vscode from "vscode";
import { ShareSession } from "./session";
import { LaunchMode } from "./launch";
import { ShareState, statusView } from "./status";

let session: ShareSession | undefined;
let statusItem: vscode.StatusBarItem | undefined;

export function activate(context: vscode.ExtensionContext): void {
  session = new ShareSession();
  session.onEnded = () => {
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
    const urls = await session!.start({
      platform: process.platform,
      arch: process.arch,
      launchMode: config.get<LaunchMode>("launchMode", "bundled"),
      pathSetting: config.get<string>("path", ""),
      binRoot: context.asAbsolutePath("bin"),
      addr: config.get<string>("addr", ""),
      hostKey: config.get<string>("hostKey", "")
    });

    setStatus("sharing");
    await vscode.env.clipboard.writeText(urls.viewer);
    await vscode.env.openExternal(vscode.Uri.parse(urls.host));
    vscode.window.showInformationMessage("termshare: sharing started — viewer link copied to clipboard.");
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);
    setStatus("error", message);
    vscode.window.showErrorMessage(`termshare: failed to start share. ${message}`);
  }
}

async function stopShare(): Promise<void> {
  if (!session?.isRunning()) {
    vscode.window.showInformationMessage("termshare is not sharing.");
    return;
  }
  await session.stop();
  setStatus("idle");
  vscode.window.showInformationMessage("termshare: share stopped.");
}

export async function deactivate(): Promise<void> {
  await session?.stop();
  session = undefined;
}
