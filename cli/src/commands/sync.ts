import * as p from "@clack/prompts";
import { loadConfig, configExists } from "../config/loader.js";
import { callCore } from "../core/client.js";

export async function syncCommand(opts: { dryRun?: boolean; delete?: boolean }): Promise<void> {
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
    args: { dryRun: !!opts.dryRun, delete: !!opts.delete },
  });

  if (!res.ok) {
    s.stop("Sync failed");
    console.error(res.error);
    process.exit(1);
  }

  const {
    pushed,
    pulled,
    conflicts,
    skipped,
    deletedLocal = 0,
    deletedRemote = 0,
    deletePending = 0,
  } = res.data as Record<string, number>;

  s.stop(opts.dryRun ? "Preview complete" : "Sync complete");

  console.log(`  ↑ push:      ${pushed}`);
  console.log(`  ↓ pull:      ${pulled}`);
  console.log(`  ⚠ conflicts: ${conflicts}`);
  console.log(`  · unchanged: ${skipped}`);
  if (deletedLocal || deletedRemote) {
    console.log(`  🗑 deleted:   ${deletedLocal + deletedRemote}`);
  }

  if (deletePending > 0) {
    console.log(`\n${deletePending} deletion(s) detected but not propagated.`);
    console.log("Re-run with --delete to archive pages / trash files.");
  }

  if (opts.dryRun) {
    console.log("\n(dry run — nothing was changed)");
  } else if (conflicts > 0) {
    console.log(`\n${conflicts} conflict(s) recorded — run \`notesync conflicts\` to review.`);
  }
}