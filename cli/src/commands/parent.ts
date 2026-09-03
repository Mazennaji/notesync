import * as p from "@clack/prompts";
import { loadConfig, saveConfig, configExists } from "../config/loader.js";
import { callCore } from "../core/client.js";

export async function parentCommand(): Promise<void> {
  const vaultPath = process.cwd();
  if (!(await configExists(vaultPath))) {
    console.error("No Notesync config here. Run `notesync init` first.");
    process.exit(1);
  }

  p.intro("Set Notion parent page");

  const input = await p.text({
    message: "Notion parent page ID (from the page URL)",
    validate: (v) => (!v || v.trim() === "" ? "Page ID is required" : undefined),
  });
  if (p.isCancel(input)) {
    p.cancel("Cancelled.");
    return;
  }

  const config = await loadConfig(vaultPath);
  const rawConfig = config as unknown as Record<string, unknown>;

  const s = p.spinner();
  s.start("Verifying parent page");
  const res = await callCore({
    command: "config.setParent",
    config: rawConfig,
    args: { parentId: input.trim() },
  });
  if (!res.ok) {
    s.stop("Verification failed");
    p.cancel(res.error ?? "Unknown error");
    return;
  }
  s.stop("Parent page verified");

  config.notionParentId = input.trim();
  await saveConfig(config);
  p.outro("Parent page saved.");
}