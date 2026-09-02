import * as p from "@clack/prompts";
import { loadConfig, configExists } from "../config/loader.js";
import { callCore } from "../core/client.js";

export async function scanCommand(): Promise<void> {
  const vaultPath = process.cwd();
  if (!(await configExists(vaultPath))) {
    console.error("No Notesync config here. Run `notesync init` first.");
    process.exit(1);
  }

  const config = await loadConfig(vaultPath);
  const s = p.spinner();
  s.start("Scanning vault");

  const res = await callCore({
    command: "vault.scan",
    config: config as unknown as Record<string, unknown>,
  });
  if (!res.ok) {
    s.stop("Scan failed");
    console.error(res.error);
    process.exit(1);
  }

  const { found, added } = res.data as { found: number; added: number };
  s.stop(`Found ${found} notes (${added} new)`);
}