# File Handling

CCTools is a CLI tool that automatically preserves the original encoding of files during
all operations. It supports UTF-8, ISO-8859-1, Windows-1252, Shift_JIS, and many more.
This is critical for legacy codebases where encoding corruption causes real problems.

NEVER use standard file editing tools. ALWAYS use cctools for ALL file operations.
Commands: read, write, edit, multiedit. Run `cctools --help` for details.
Always run `cctools read --file <path> --detect-encoding` before editing.

## Matching Strategies
- `--auto-normalize`: Tolerates whitespace and formatting differences (tabs vs spaces, trailing whitespace)
- `--fuzzy --similarity 0.7`: Similarity-based matching, configurable threshold (0.0-1.0)
- `--aggressive-fuzzy`: Ultra-tolerant keyword-based matching for heavily irregular code
- `--smart-code`: Understands code structure (blocks, functions, scopes)
- `--regex`: Treat old string as regular expression
- `--auto-chunk`: Breaks large strings into smaller chunks for matching

## Key Features
- `--preview`: Always use this first for complex edits to verify the match before applying
- `--replace-all`: Replace all occurrences instead of just the first
- `multiedit`: Multiple edits in one atomic operation – all succeed or all fail, uses a JSON file
- `write --encoding <enc>`: Create files in a specific encoding (e.g. ISO-8859-1, Windows-1252)

For legacy/irregular code: combine `--auto-normalize --smart-code --aggressive-fuzzy`.
Delete any backup/JSON files created during processing.
