import { loadConfig, configExists } from "../config/loader.js";
import { callCore } from "../core/client.js";

export async function conflictsCommand(): Promise<void> {
  const vaultPath = process.cwd();
  if (!(await configExists(vaultPath))) {
    console.error("No Notesync config here. Run `notesync init` first.");
    process.exit(1);
  }

  const config = await loadConfig(vaultPath);
  const res = await callCore({
    command: "conflicts.list",
    config: config as unknown as Record<string, unknown>,
  });

  if (!res.ok) {
    console.error(res.error);
    process.exit(1);
  }

  const { conflicts } = res.data as {
    conflicts?: { id: string; noteId: string; note: string; detectedAt: string }[];
  };

  if (!conflicts || conflicts.length === 0) {
    console.log("No unresolved conflicts. ✓");
    return;
  }

  console.log(`${conflicts.length} unresolved conflict(s):\n`);
  for (const c of conflicts) {
    console.log(`  ⚠ ${c.note}   (detected ${c.detectedAt})`);
  }
  console.log(`\nRun \`notesync resolve\` to resolve them interactively.`);
}