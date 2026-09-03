import * as p from "@clack/prompts";
import { loadConfig, configExists } from "../config/loader.js";
import { callCore } from "../core/client.js";

export async function pullCommand(): Promise<void> {
  const vaultPath = process.cwd();
  if (!(await configExists(vaultPath))) {
    console.error("No Notesync config here. Run `notesync init` first.");
    process.exit(1);
  }

  const config = await loadConfig(vaultPath);
  const s = p.spinner();
  s.start("Pulling note content from Notion");

  const res = await callCore({
    command: "notion.pull",
    config: config as unknown as Record<string, unknown>,
  });

  if (!res.ok) {
    s.stop("Pull failed");
    console.error(res.error);
    process.exit(1);
  }

  const { pulled, skipped } = res.data as { pulled: number; skipped: number };
  s.stop(`Pulled ${pulled} note(s), skipped ${skipped} unchanged`);
}