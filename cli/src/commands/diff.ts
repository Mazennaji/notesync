import * as p from "@clack/prompts";
import { loadConfig, configExists } from "../config/loader.js";
import { callCore } from "../core/client.js";

const LABELS: Record<string, string> = {
  skip: "unchanged",
  push: "→ push to Notion",
  pull: "← pull from Notion",
  conflict: "⚠ conflict",
  new_local: "＋ new (never synced)",
};

export async function diffCommand(): Promise<void> {
  const vaultPath = process.cwd();
  if (!(await configExists(vaultPath))) {
    console.error("No Notesync config here. Run `notesync init` first.");
    process.exit(1);
  }

  const config = await loadConfig(vaultPath);
  const s = p.spinner();
  s.start("Comparing vault and Notion");

  const res = await callCore({
    command: "sync.diff",
    config: config as unknown as Record<string, unknown>,
  });

  if (!res.ok) {
    s.stop("Diff failed");
    console.error(res.error);
    process.exit(1);
  }
  s.stop("Comparison complete");

  const data = res.data as {
    decisions: { note: string; action: string }[];
    counts: Record<string, number>;
  };

  if (data.decisions.length === 0) {
    console.log("No linked notes to compare.");
    return;
  }

  for (const d of data.decisions) {
    console.log(`  ${LABELS[d.action] ?? d.action}   ${d.note}`);
  }

  const summary = Object.entries(data.counts)
    .map(([k, v]) => `${v} ${LABELS[k] ?? k}`)
    .join(", ");
  console.log(`\n${summary}`);
}