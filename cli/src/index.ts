import { Command } from "commander";
import { callCore } from "./core/client.js";
import { initCommand } from "./commands/init.js";
import { authCommand, scanCommand } from "./commands/scan.js";
import { loadConfig, configExists } from "./config/loader.js"
import { pagesCommand } from "./commands/pages.js";
import { parentCommand } from "./commands/parent.js";
import { createCommand } from "./commands/create.js";
import { pushCommand } from "./commands/push.js";
import { diffCommand } from "./commands/diff.js";
import { pullCommand } from "./commands/pull.js";
import { syncCommand } from "./commands/sync.js";
import { conflictsCommand } from "./commands/conflicts.js";
import { resolveCommand } from "./commands/resolve.js";
import { historyCommand } from "./commands/history.js";
import { watchCommand } from "./commands/watch.js";

const program = new Command();

program
  .name("notesync")
  .description("Bidirectional sync between Notion and Obsidian")
  .version("0.1.0");

program.command("init").description("Initialize a Notesync configuration")
  .action(initCommand);
program.command("auth")
  .description("Configure Notion authentication")
  .action(authCommand);
program.command("status").description("Display synchronization status")
  .action(async () => {
    const vaultPath = process.cwd();
    if (!(await configExists(vaultPath))) {
      console.error("No Notesync config here. Run `notesync init` first.");
      process.exit(1);
    }
    const config = await loadConfig(vaultPath);
    const res = await callCore({
      command: "status",
      config: config as unknown as Record<string, unknown>,
    });
    if (!res.ok) { console.error(res.error); process.exit(1); }
    const d = res.data as { vaultPath: string; syncMode: string; notes: number };
    console.log(`Vault:  ${d.vaultPath}`);
    console.log(`Mode:   ${d.syncMode}`);
    console.log(`Notes:  ${d.notes}`);
  });
program.command("sync")
  .description("Synchronize both directions")
  .option("--dry-run", "Preview changes without applying them")
  .option("--delete", "Propagate deletions (archive pages / trash files)")
  .action(syncCommand);

program.command("scan")
  .description("Discover Markdown notes in the vault")
  .action(scanCommand);

program.command("logout")
  .description("Remove stored Notion credentials")
  .action(async () => {
    const res = await callCore({ command: "auth.logout" });
    if (!res.ok) { console.error(res.error); process.exit(1); }
    console.log("Logged out. Notion token removed.");
  });

program.command("pages")
  .description("Discover Notion pages and link them to vault notes")
  .action(pagesCommand);

program.command("parent")
  .description("Set the Notion parent page for new notes")
  .action(parentCommand);

program.command("create")
  .description("Create Notion pages for notes not yet linked")
  .action(createCommand);
  
program.command("push")
  .description("Push local note content to Notion")
  .action(pushCommand);

program.command("diff")
  .description("Show what would be synchronized (read-only)")
  .action(diffCommand);

program.command("pull")
  .description("Pull note content from Notion into the vault")
  .action(pullCommand);

program.command("conflicts")
  .description("List unresolved synchronization conflicts")
  .action(conflictsCommand);

program.command("resolve")
  .description("Interactively resolve conflicts")
  .action(resolveCommand);

program.command("history")
  .description("Show previous synchronization operations")
  .option("-n, --limit <number>", "How many entries to show", "20")
  .action(historyCommand);

program.command("watch")
  .description("Watch the vault and synchronize automatically")
  .option("-i, --interval <seconds>", "Also poll Notion every N seconds for remote changes")
  .action(watchCommand);

program
  .option("--verbose", "Show internal core logs")
  .hook("preAction", (thisCommand) => {
    if (thisCommand.opts().verbose) {
      process.env.NOTESYNC_VERBOSE = "1";
    }
  });

program.parse();