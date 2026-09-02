import { readFile, writeFile, mkdir, stat } from "node:fs/promises";
import { join } from "node:path";
import { NotesyncConfig, CONFIG_DIR, CONFIG_FILE } from "./schema.js";

function configPath(vaultPath: string): string {
  return join(vaultPath, CONFIG_DIR, CONFIG_FILE);
}

export async function configExists(vaultPath: string): Promise<boolean> {
  try {
    await stat(configPath(vaultPath));
    return true;
  } catch {
    return false;
  }
}

export async function loadConfig(vaultPath: string): Promise<NotesyncConfig> {
  const raw = await readFile(configPath(vaultPath), "utf8");
  return JSON.parse(raw) as NotesyncConfig;
}

export async function saveConfig(config: NotesyncConfig): Promise<void> {
  const dir = join(config.vaultPath, CONFIG_DIR);
  await mkdir(dir, { recursive: true });
  await writeFile(
    join(dir, CONFIG_FILE),
    JSON.stringify(config, null, 2) + "\n",
    "utf8",
  );
}