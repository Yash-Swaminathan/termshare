import { spawn, ChildProcess } from "node:child_process";
import * as readline from "node:readline";
import { parseShareJSON, ShareURLs } from "./parse";
import { resolveSpawn, ResolveOptions } from "./launch";

const START_TIMEOUT_MS = 10_000;

export class ShareSession {
  private child?: ChildProcess;
  private urls?: ShareURLs;

  // Called when the termshare process exits on its own (e.g. the shared
  // shell ended), as opposed to stop() being invoked.
  onEnded?: () => void;

  isRunning(): boolean {
    return this.child !== undefined;
  }

  get current(): ShareURLs | undefined {
    return this.urls;
  }

  async start(opts: ResolveOptions): Promise<ShareURLs> {
    if (this.child) {
      throw new Error("a share is already running");
    }

    const plan = resolveSpawn(opts);
    const child = spawn(plan.command, plan.args, {
      cwd: plan.cwd,
      stdio: ["ignore", "pipe", "pipe"]
    });
    this.child = child;

    let stderr = "";
    child.stderr?.on("data", (chunk) => {
      stderr += chunk.toString();
    });

    try {
      const urls = await this.readFirstJSON(child, () => stderr);
      this.urls = urls;
      child.once("exit", () => {
        if (this.child === child) {
          this.child = undefined;
          this.urls = undefined;
          this.onEnded?.();
        }
      });
      return urls;
    } catch (err) {
      this.forceKill();
      throw err;
    }
  }

  private readFirstJSON(child: ChildProcess, stderr: () => string): Promise<ShareURLs> {
    return new Promise<ShareURLs>((resolve, reject) => {
      const rl = readline.createInterface({ input: child.stdout! });
      const timer = setTimeout(() => {
        cleanup();
        reject(new Error(`termshare did not print a share URL within ${START_TIMEOUT_MS}ms. ${stderr().trim()}`));
      }, START_TIMEOUT_MS);

      const cleanup = () => {
        clearTimeout(timer);
        rl.close();
        child.removeListener("exit", onExit);
        child.removeListener("error", onError);
      };

      const onLine = (line: string) => {
        const trimmed = line.trim();
        if (!trimmed.startsWith("{")) {
          return;
        }
        try {
          const urls = parseShareJSON(trimmed);
          cleanup();
          resolve(urls);
        } catch (err) {
          cleanup();
          reject(err instanceof Error ? err : new Error(String(err)));
        }
      };

      const onExit = (code: number | null) => {
        cleanup();
        reject(new Error(`termshare exited (code ${code}) before printing a share URL. ${stderr().trim()}`));
      };

      const onError = (err: Error) => {
        cleanup();
        reject(new Error(`failed to launch termshare: ${err.message}`));
      };

      rl.on("line", onLine);
      child.once("exit", onExit);
      child.once("error", onError);
    });
  }

  async stop(): Promise<void> {
    const child = this.child;
    if (!child) {
      return;
    }
    this.child = undefined;
    this.urls = undefined;

    if (child.exitCode !== null || child.signalCode !== null) {
      return;
    }
    await new Promise<void>((resolve) => {
      child.once("exit", () => resolve());
      child.kill();
      setTimeout(() => {
        if (child.exitCode === null && child.signalCode === null) {
          child.kill("SIGKILL");
        }
        resolve();
      }, 2_000);
    });
  }

  private forceKill(): void {
    const child = this.child;
    this.child = undefined;
    this.urls = undefined;
    if (child && child.exitCode === null && child.signalCode === null) {
      child.kill("SIGKILL");
    }
  }
}
