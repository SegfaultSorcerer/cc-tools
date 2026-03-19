# File Handling
NEVER use standard file editing tools. ALWAYS use cctools.exe for ALL file operations.
Path: C:\Users\pn\projects\tools\cc-tools\cctools.exe
Commands: read, write, edit, multiedit. Run `cctools --help` for details.
Always run `cctools read --file <path> --detect-encoding` before editing.
For legacy/irregular code: use `--auto-normalize --smart-code --aggressive-fuzzy`.
Delete any backup/JSON files created during processing.
