import { resolve } from "node:path";
import * as p from "@clack/prompts";
import { NotesyncConfig, CONFIG_VERSION } from "../config/schema.js";
import { configExists, saveConfig } from "../config/loader.js";
import { isDirectory } from "../config/validate.js";
import { callCore } from "../core/client.js";

export async function initCommand(): Promise<void> {
  p.intro("Initialize Notesync");

  const vaultInput = await p.text({
    message: "Path to your Obsidian vault",
    placeholder: process.cwd(),
    defaultValue: process.cwd(),
    validate: (v) => (v.trim() === "" ? "Path is required" : undefined),
  });
  if (p.isCancel(vaultInput)) return p.cancel("Cancelled.");

  const vaultPath = resolve(vaultInput);

  if (!(await isDirectory(vaultPath))) {
    return p.cancel(`Not a directory: ${vaultPath}`);
  }
  if (await configExists(vaultPath)) {
    const overwrite = await p.confirm({ message: "Config already exists. Overwrite?" });
    if (p.isCancel(overwrite) || !overwrite) return p.cancel("Cancelled.");
  }

  const syncMode = await p.select({
    message: "Sync mode",
    options: [
      { value: "manual", label: "Manual — sync when you run it" },
      { value: "watch", label: "Watch — sync automatically on changes" },
    ],
  });
  if (p.isCancel(syncMode)) return p.cancel("Cancelled.");

  const config: NotesyncConfig = {
    version: CONFIG_VERSION,
    vaultPath,
    notionParentId: null,
    syncMode: syncMode as "manual" | "watch",
  };

  const res = await callCore({
    command: "config.validate",
    config: config as unknown as Record<string, unknown>,
  });
  if (!res.ok) return p.cancel(`Core rejected config: ${res.error}`);

  await saveConfig(config);
  p.outro(`Initialized at ${vaultPath}/.notesync/config.json`);
}