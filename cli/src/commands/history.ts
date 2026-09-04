import { loadConfig, configExists } from "../config/loader.js";
import { callCore } from "../core/client.js";

const ICONS: Record<string, string> = {
  push: "↑",
  pull: "↓",
  sync: "⇅",
  resolve: "✓",
  merge: "⛙",
  delete: "🗑",
};

const STATUS: Record<string, string> = {
  success: "ok",
  error: "ERR",
  conflict: "⚠",
  partial: "~",
};

export async function historyCommand(opts: { limit?: string }): Promise<void> {
  const vaultPath = process.cwd();
  if (!(await configExists(vaultPath))) {
    console.error("No Notesync config here. Run `notesync init` first.");
    process.exit(1);
  }

  const config = await loadConfig(vaultPath);
  const limit = opts.limit ? Number(opts.limit) : 20;

  const res = await callCore({
    command: "history.list",
    config: config as unknown as Record<string, unknown>,
    args: { limit },
  });

  if (!res.ok) {
    console.error(res.error);
    process.exit(1);
  }

  const { entries } = res.data as {
    entries?: {
      note: string;
      operation: string;
      direction: string;
      status: string;
      error: string;
      at: string;
    }[];
  };

  if (!entries || entries.length === 0) {
    console.log("No sync history yet.");
    return;
  }

  console.log(`Last ${entries.length} operation(s):\n`);
  for (const e of entries) {
    const icon = ICONS[e.operation] ?? "·";
    const status = STATUS[e.status] ?? e.status;
    const when = e.at.replace("T", " ").replace("Z", "");
    let line = `  ${when}  ${icon} ${e.operation.padEnd(7)} ${status.padEnd(3)}  ${e.note}`;
    if (e.error) line += `\n      └ ${e.error}`;
    console.log(line);
  }
}