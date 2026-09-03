import * as p from "@clack/prompts";
import { loadConfig, configExists } from "../config/loader.js";
import { callCore } from "../core/client.js";

export async function createCommand(): Promise<void> {
  const vaultPath = process.cwd();
  if (!(await configExists(vaultPath))) {
    console.error("No Notesync config here. Run `notesync init` first.");
    process.exit(1);
  }

  const config = await loadConfig(vaultPath);
  const s = p.spinner();
  s.start("Creating Notion pages for unlinked notes");

  const res = await callCore({
    command: "notion.create",
    config: config as unknown as Record<string, unknown>,
  });

  if (!res.ok) {
    s.stop("Creation failed");
    console.error(res.error);
    process.exit(1);
  }

  const { created } = res.data as { created: number };
  s.stop(`Created ${created} Notion page(s)`);
}