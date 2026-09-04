import { join } from "node:path";
import chokidar from "chokidar";
import * as p from "@clack/prompts";
import { loadConfig, configExists } from "../config/loader.js";
import { callCore } from "../core/client.js";

async function runSync(rawConfig: Record<string, unknown>): Promise<void> {
  const res = await callCore({
    command: "notion.sync",
    config: rawConfig,
    args: { dryRun: false, delete: false },
  });

  const stamp = new Date().toLocaleTimeString();
  if (!res.ok) {
    console.log(`  [${stamp}] sync failed: ${res.error}`);
    return;
  }
  const { pushed, pulled, conflicts, skipped } = res.data as Record<string, number>;
  const parts: string[] = [];
  if (pushed) parts.push(`↑${pushed}`);
  if (pulled) parts.push(`↓${pulled}`);
  if (conflicts) parts.push(`⚠${conflicts}`);
  const summary = parts.length ? parts.join(" ") : "no changes";
  console.log(`  [${stamp}] ${summary}${skipped ? `  (${skipped} unchanged)` : ""}`);
  if (conflicts) {
    console.log(`  [${stamp}] ${conflicts} conflict(s) — run \`notesync resolve\``);
  }
}

export async function watchCommand(opts: { interval?: string }): Promise<void> {
  const vaultPath = process.cwd();
  if (!(await configExists(vaultPath))) {
    console.error("No Notesync config here. Run `notesync init` first.");
    process.exit(1);
  }

  const config = await loadConfig(vaultPath);
  const rawConfig = config as unknown as Record<string, unknown>;

  const pollSeconds = opts.interval ? Number(opts.interval) : 0;

  p.intro("Notesync watch mode");
  console.log(`  Watching: ${vaultPath}`);
  if (pollSeconds > 0) {
    console.log(`  Polling Notion every ${pollSeconds}s`);
  }
  console.log("  Press Ctrl+C to stop.\n");

  let timer: NodeJS.Timeout | null = null;
  let syncing = false;

  const scheduleSync = () => {
    if (timer) clearTimeout(timer);
    timer = setTimeout(async () => {
      if (syncing) return;
      syncing = true;
      try {
        await runSync(rawConfig);
      } finally {
        syncing = false;
      }
    }, 800);
  };

  const watcher = chokidar.watch(join(vaultPath, "**", "*.md"), {
    ignored: [/(^|[\/\\])\../, join(vaultPath, ".notesync", "**")],
    ignoreInitial: true,
    persistent: true,
  });

  watcher
    .on("add", scheduleSync)
    .on("change", scheduleSync)
    .on("unlink", scheduleSync);

  let poller: NodeJS.Timeout | null = null;
  if (pollSeconds > 0) {
    poller = setInterval(scheduleSync, pollSeconds * 1000);
  }

  await runSync(rawConfig);

  process.on("SIGINT", async () => {
    console.log("\nStopping watch mode…");
    await watcher.close();
    if (poller) clearInterval(poller);
    process.exit(0);
  });
}