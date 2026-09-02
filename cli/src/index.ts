import { Command } from "commander";
import { callCore } from "./core/client.js";
import { initCommand } from "./commands/init.js";
import { scanCommand } from "./commands/scan.js";
import { loadConfig, configExists } from "./config/loader.js"

const program = new Command();

program
  .name("notesync")
  .description("Bidirectional sync between Notion and Obsidian")
  .version("0.1.0");

program.command("init").description("Initialize a Notesync configuration")
  .action(initCommand);
program.command("auth").description("Configure Notion authentication")
  .action(() => console.log("auth: not yet implemented"));
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
program.command("sync").description("Synchronize both directions")
  .option("--dry-run", "Preview without modifying data")
  .action((opts) => console.log("sync: not yet implemented", opts));

program.command("scan")
  .description("Discover Markdown notes in the vault")
  .action(scanCommand);

program.parse();