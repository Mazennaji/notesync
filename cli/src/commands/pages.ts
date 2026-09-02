import * as p from "@clack/prompts";
import { loadConfig, configExists } from "../config/loader.js";
import { callCore } from "../core/client.js";

export async function pagesCommand(): Promise<void> {
  const vaultPath = process.cwd();
  if (!(await configExists(vaultPath))) {
    console.error("No Notesync config here. Run `notesync init` first.");
    process.exit(1);
  }

  const config = await loadConfig(vaultPath);
  const s = p.spinner();
  s.start("Discovering Notion pages");

  const res = await callCore({
    command: "notion.pages",
    config: config as unknown as Record<string, unknown>,
  });

  if (!res.ok) {
    s.stop("Discovery failed");
    console.error(res.error);
    process.exit(1);
  }

  const { found, linked } = res.data as { found: number; linked: number };
  s.stop(`Found ${found} Notion pages, linked ${linked} to vault notes`);
}