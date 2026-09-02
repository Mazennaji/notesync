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
  if (p.isCancel(vaultInput)) {
    p.cancel("Cancelled.");
    return;
  }

  const vaultPath = resolve(vaultInput);

  if (!(await isDirectory(vaultPath))) {
    p.cancel(`Not a directory: ${vaultPath}`);
    return;
  }

  if (await configExists(vaultPath)) {
    const overwrite = await p.confirm({
      message: "Config already exists. Overwrite?",
    });
    if (p.isCancel(overwrite) || !overwrite) {
      p.cancel("Cancelled.");
      return;
    }
  }

  const syncMode = await p.select({
    message: "Sync mode",
    options: [
      { value: "manual", label: "Manual — sync when you run it" },
      { value: "watch", label: "Watch — sync automatically on changes" },
    ],
  });
  if (p.isCancel(syncMode)) {
    p.cancel("Cancelled.");
    return;
  }

  const config: NotesyncConfig = {
    version: CONFIG_VERSION,
    vaultPath,
    notionParentId: null,
    syncMode: syncMode as "manual" | "watch",
  };

  const rawConfig = config as unknown as Record<string, unknown>;

  const s = p.spinner();
  s.start("Validating configuration");
  const validateRes = await callCore({ command: "config.validate", config: rawConfig });
  if (!validateRes.ok) {
    s.stop("Validation failed");
    p.cancel(`Core rejected config: ${validateRes.error}`);
    return;
  }

  await saveConfig(config);

  s.message("Creating state database");
  const dbRes = await callCore({ command: "db.init", config: rawConfig });
  if (!dbRes.ok) {
    s.stop("Database setup failed");
    p.cancel(`Failed to create state DB: ${dbRes.error}`);
    return;
  }
  s.stop("Ready");

  const dbPath = (dbRes.data as { dbPath: string }).dbPath;
  p.outro(`Initialized:\n  config: ${vaultPath}/.notesync/config.json\n  state:  ${dbPath}`);
}