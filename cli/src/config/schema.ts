export interface NotesyncConfig {
  version: string;
  vaultPath: string;
  notionParentId: string | null;
  syncMode: "manual" | "watch";
}

export const CONFIG_VERSION = "1";
export const CONFIG_DIR = ".notesync";
export const CONFIG_FILE = "config.json";