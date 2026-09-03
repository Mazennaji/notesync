import * as p from "@clack/prompts";
import { loadConfig, configExists } from "../config/loader.js";
import { callCore } from "../core/client.js";

function preview(md: string, lines = 15): string {
  const parts = md.split("\n");
  const out = parts.slice(0, lines).join("\n");
  return parts.length > lines ? out + "\n  …" : out;
}

export async function resolveCommand(): Promise<void> {
  const vaultPath = process.cwd();
  if (!(await configExists(vaultPath))) {
    console.error("No Notesync config here. Run `notesync init` first.");
    process.exit(1);
  }

  const config = await loadConfig(vaultPath);
  const rawConfig = config as unknown as Record<string, unknown>;

  const listRes = await callCore({ command: "conflicts.list", config: rawConfig });
  if (!listRes.ok) {
    console.error(listRes.error);
    process.exit(1);
  }
  const { conflicts } = listRes.data as {
    conflicts?: { id: string; noteId: string; note: string }[];
  };

  if (!conflicts || conflicts.length === 0) {
    console.log("No unresolved conflicts. \u2713");
    return;
  }

  p.intro(`Resolving ${conflicts.length} conflict(s)`);

  let resolved = 0;
  let skipped = 0;

  for (const c of conflicts) {
    const prevRes = await callCore({
      command: "conflicts.preview",
      config: rawConfig,
      args: { noteId: Number(c.noteId) },
    });

    if (!prevRes.ok) {
      p.log.warn(`Skipping ${c.note}: ${prevRes.error}`);
      continue;
    }

    const { local, remote } = prevRes.data as { local: string; remote: string };

    p.log.step(`Conflict: ${c.note}`);
    p.log.message(`LOCAL:\n${preview(local)}`);
    p.log.message(`REMOTE (Notion):\n${preview(remote)}`);

    const choice = await p.select({
      message: "How should this be resolved?",
      options: [
        { value: "local", label: "Keep local — overwrite Notion" },
        { value: "remote", label: "Keep remote — overwrite local file" },
        { value: "skip", label: "Skip — decide later" },
      ],
    });

    if (p.isCancel(choice)) {
      p.cancel("Stopped. Remaining conflicts left unresolved.");
      return;
    }

    const applyRes = await callCore({
      command: "conflicts.resolve",
      config: rawConfig,
      args: { conflictId: Number(c.id), noteId: Number(c.noteId), choice },
    });

    if (!applyRes.ok) {
      p.log.error(`Failed to resolve ${c.note}: ${applyRes.error}`);
      continue;
    }

    if (choice === "skip") {
      skipped++;
    } else {
      resolved++;
    }
  }

  p.outro(`Resolved ${resolved}, skipped ${skipped}.`);
}