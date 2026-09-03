import * as p from "@clack/prompts";
import { loadConfig, configExists } from "../config/loader.js";
import { callCore } from "../core/client.js";

export async function syncCommand(): Promise<void> {
  const vaultPath = process.cwd();
  if (!(await configExists(vaultPath))) {
    console.error("No Notesync config here. Run `notesync init` first.");
    process.exit(1);
  }

  const config = await loadConfig(vaultPath);
  const s = p.spinner();
  s.start("Synchronizing vault and Notion");

  const res = await callCore({
    command: "notion.sync",
    config: config as unknown as Record<string, unknown>,
  });

  if (!res.ok) {
    s.stop("Sync failed");
    console.error(res.error);
    process.exit(1);
  }

  const { pushed, pulled, conflicts, skipped } = res.data as {
    pushed: number; pulled: number; conflicts: number; skipped: number;
  };
  s.stop("Sync complete");

  console.log(`  ↑ pushed:    ${pushed}`);
  console.log(`  ↓ pulled:    ${pulled}`);
  console.log(`  ⚠ conflicts: ${conflicts}`);
  console.log(`  · unchanged: ${skipped}`);

  if (conflicts > 0) {
    console.log(`\n${conflicts} conflict(s) recorded — run \`notesync conflicts\` to review.`);
  }
}