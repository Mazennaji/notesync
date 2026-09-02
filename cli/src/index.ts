import { Command } from "commander";
import { callCore } from "./core/client.js";
import { initCommand } from "./commands/init.js";

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
    const res = await callCore({ command: "status" });
    if (!res.ok) { program.error(`Error: ${res.error}`); }
    console.log(res.data);
  });
program.command("sync").description("Synchronize both directions")
  .option("--dry-run", "Preview without modifying data")
  .action((opts) => console.log("sync: not yet implemented", opts));

program.parse();