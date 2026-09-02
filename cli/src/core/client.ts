import { spawn } from "node:child_process";
import { env } from "node:process";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

const __dirname = dirname(fileURLToPath(import.meta.url));

export interface CoreRequest {
  command: string;
  args?: Record<string, unknown>;
  config?: Record<string, unknown>;
}

export interface CoreResponse<T = unknown> {
  ok: boolean;
  data?: T;
  error?: string;
}

const CORE_BIN = env.NOTESYNC_CORE
  ?? join(__dirname, "../../../core/notesync-core");

export function callCore<T = unknown>(req: CoreRequest): Promise<CoreResponse<T>> {
  return new Promise((resolve, reject) => {
    const proc = spawn(CORE_BIN, [], { stdio: ["pipe", "pipe", "pipe"] });

    let stdout = "";
    let stderr = "";
    proc.stdout.on("data", (d) => (stdout += d));
    proc.stderr.on("data", (d) => (stderr += d));

    proc.on("error", reject);
    proc.on("close", (code) => {
      if (code !== 0 && !stdout) {
        return reject(new Error(`core exited ${code}: ${stderr}`));
      }
      try {
        resolve(JSON.parse(stdout));
      } catch {
        reject(new Error(`invalid response from core: ${stdout || stderr}`));
      }
    });

    proc.stdin.write(JSON.stringify(req));
    proc.stdin.end();
  });
}