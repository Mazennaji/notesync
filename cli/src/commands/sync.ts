import * as p from "@clack/prompts";
import { loadConfig, configExists } from "../config/loader.js";
import { callCore } from "../core/client.js";

export async function syncCommand(opts: { dryRun?: boolean }): Promise<void> {
  const vaultPath = process.cwd();
  if (!(await configExists(vaultPath))) {
    console.error("No Notesync config here. Run `notesync init` first.");
    process.exit(1);
  }

  const config = await loadConfig(vaultPath);
  const s = p.spinner();
  s.start(opts.dryRun ? "Previewing sync" : "Synchronizing vault and Notion");

  const res = await callCore({
    command: "notion.sync",
    config: config as unknown as Record<string, unknown>,
    args: { dryRun: !!opts.dryRun },
  });

  if (!res.ok) {
    s.stop("Sync failed");
    console.error(res.error);
    process.exit(1);
  }

  const { pushed, pulled, conflicts, skipped } = res.data as {
    pushed: number; pulled: number; conflicts: number; skipped: number;
  };
  s.stop(opts.dryRun ? "Preview complete" : "Sync complete");

  console.log(`  ↑ push:      ${pushed}`);
  console.log(`  ↓ pull:      ${pulled}`);
  console.log(`  ⚠ conflicts: ${conflicts}`);
  console.log(`  · unchanged: ${skipped}`);

  if (opts.dryRun) {
    console.log("\n(dry run — nothing was changed)");
  } else if (conflicts > 0) {
    console.log(`\n${conflicts} conflict(s) recorded — run \`notesync conflicts\` to review.`);
  }
}