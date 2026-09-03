import * as p from "@clack/prompts";
import { loadConfig, configExists } from "../config/loader.js";
import { callCore } from "../core/client.js";

export async function pushCommand(): Promise<void> {
  const vaultPath = process.cwd();
  if (!(await configExists(vaultPath))) {
    console.error("No Notesync config here. Run `notesync init` first.");
    process.exit(1);
  }

  const config = await loadConfig(vaultPath);
  const s = p.spinner();
  s.start("Pushing note content to Notion");

  const res = await callCore({
    command: "notion.push",
    config: config as unknown as Record<string, unknown>,
  });

  if (!res.ok) {
    s.stop("Push failed");
    console.error(res.error);
    process.exit(1);
  }

  const { pushed } = res.data as { pushed: number };
  s.stop(`Pushed content for ${pushed} note(s)`);
}