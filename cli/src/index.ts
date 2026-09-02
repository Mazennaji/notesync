import { Command } from "commander";

const program = new Command();

program
  .name("notesync")
  .description("Bidirectional sync between Notion and Obsidian")
  .version("0.1.0");

program.command("init").description("Initialize a Notesync configuration")
  .action(() => console.log("init: not yet implemented"));
program.command("auth").description("Configure Notion authentication")
  .action(() => console.log("auth: not yet implemented"));
program.command("status").description("Display synchronization status")
  .action(() => console.log("status: not yet implemented"));
program.command("sync").description("Synchronize both directions")
  .option("--dry-run", "Preview without modifying data")
  .action((opts) => console.log("sync: not yet implemented", opts));

program.parse();