import * as vscode from "vscode";
import { ShareSession } from "./session";

let session: ShareSession | undefined;

export function activate(context: vscode.ExtensionContext): void {
  session = new ShareSession();

  context.subscriptions.push(
    vscode.commands.registerCommand("termshare.startShare", startShare),
    vscode.commands.registerCommand("termshare.stopShare", stopShare)
  );
}

async function startShare(): Promise<void> {
  if (session?.isRunning()) {
    vscode.window.showInformationMessage("termshare is already sharing.");
    return;
  }

  const pathSetting = vscode.workspace.getConfiguration("termshare").get<string>("path", "");
  const workspaceRoot = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;

  try {
    const urls = await session!.start({
      platform: process.platform,
      pathSetting,
      workspaceRoot
    });

    await vscode.env.clipboard.writeText(urls.viewer);
    await vscode.env.openExternal(vscode.Uri.parse(urls.host));

    const openHost = "Open host";
    const choice = await vscode.window.showInformationMessage(
      "termshare: viewer link copied to clipboard.",
      openHost
    );
    if (choice === openHost) {
      await vscode.env.openExternal(vscode.Uri.parse(urls.host));
    }
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);
    vscode.window.showErrorMessage(`termshare: failed to start share. ${message}`);
  }
}

async function stopShare(): Promise<void> {
  if (!session?.isRunning()) {
    vscode.window.showInformationMessage("termshare is not sharing.");
    return;
  }
  await session.stop();
  vscode.window.showInformationMessage("termshare: share stopped.");
}

export async function deactivate(): Promise<void> {
  await session?.stop();
  session = undefined;
}
