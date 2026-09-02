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

export async function authCommand(): Promise<void> {
  p.intro("Authenticate with Notion");

  p.note(
    "Create an integration at https://www.notion.so/my-integrations\n" +
      "then copy the Internal Integration Secret.",
    "How to get a token",
  );

  const token = await p.password({
    message: "Notion integration token",
    validate: (v) => (!v || v.trim() === "" ? "Token is required" : undefined),
  });
  if (p.isCancel(token)) {
    p.cancel("Cancelled.");
    return;
  }

  const s = p.spinner();
  s.start("Verifying with Notion");

  const res = await callCore({
    command: "auth.store",
    args: { token },
  });

  if (!res.ok) {
    s.stop("Authentication failed");
    p.cancel(res.error ?? "Unknown error");
    return;
  }

  const { integration } = res.data as { integration: string };
  s.stop(`Authenticated as "${integration}"`);
  p.outro("Token saved to your OS credential store.");
}