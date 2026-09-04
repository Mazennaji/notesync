const COMMANDS = [
  "init", "auth", "logout", "parent", "scan", "pages", "create",
  "status", "diff", "push", "pull", "sync", "conflicts", "resolve",
  "history", "watch", "completion",
];

export function completionCommand(opts: { shell?: string }): void {
  const shell = opts.shell ?? "bash";
  const cmds = COMMANDS.join(" ");

  if (shell === "bash") {
    console.log(`# notesync bash completion — add to ~/.bashrc:
#   source <(notesync completion --shell bash)
_notesync() {
  local cur="\${COMP_WORDS[COMP_CWORD]}"
  COMPREPLY=( $(compgen -W "${cmds}" -- "$cur") )
}
complete -F _notesync notesync`);
  } else if (shell === "zsh") {
    console.log(`# notesync zsh completion — add to ~/.zshrc:
#   source <(notesync completion --shell zsh)
_notesync() { compadd ${cmds} }
compdef _notesync notesync`);
  } else if (shell === "powershell") {
    console.log(`# notesync PowerShell completion — add to $PROFILE:
Register-ArgumentCompleter -Native -CommandName notesync -ScriptBlock {
  param($wordToComplete, $commandAst, $cursorPosition)
  @(${COMMANDS.map((c) => `'${c}'`).join(", ")}) |
    Where-Object { $_ -like "$wordToComplete*" } |
    ForEach-Object { [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_) }
}`);
  } else {
    console.error(`Unsupported shell: ${shell} (use bash, zsh, or powershell)`);
    process.exit(1);
  }
}