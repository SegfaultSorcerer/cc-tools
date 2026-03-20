# CCTools — Project Setup Guide

This guide explains how to enforce cctools usage in your project with Claude Code.

## Prerequisites

- [Claude Code](https://claude.ai/code) installed
- `cctools` binary in your PATH or project directory
  - Download from [Releases](https://github.com/SegfaultSorcerer/cc-tools/releases)

## Setup

### Option A: Team-shared (committed to git)

This ensures everyone on the team uses cctools automatically.

1. **Copy the hook script** into your project:

```bash
mkdir -p scripts/hooks
# Linux/macOS:
cp enforce-cctools.sh scripts/hooks/
chmod +x scripts/hooks/enforce-cctools.sh
# Windows:
cp enforce-cctools.cmd scripts/hooks/
```

2. **Create `.claude/settings.json`** in your project root:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "bash scripts/hooks/enforce-cctools.sh"
          }
        ]
      }
    ]
  }
}
```

On Windows, replace the command with:
```json
"command": "scripts\\hooks\\enforce-cctools.cmd"
```

3. **Add CLAUDE.md** to your project root. Copy from [prompt-short.md](../prompt-short.md) or use:

```markdown
# File Handling

NEVER use standard file editing tools. ALWAYS use cctools for ALL file operations.
Commands: read, write, edit, multiedit. Run `cctools --help` for details.
Always run `cctools read --file <path> --detect-encoding` before editing.
```

4. **Commit** all three (hook script, settings, CLAUDE.md) so the team gets it automatically.

### Option B: Personal (not committed)

For individual use without affecting the team.

1. **Copy the hook script** anywhere (e.g., `~/.claude/hooks/enforce-cctools.sh`)

2. **Create `.claude/settings.local.json`** in your project root (this file is typically gitignored):

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "bash ~/.claude/hooks/enforce-cctools.sh"
          }
        ]
      }
    ]
  }
}
```

## How it works

The PreToolUse hook intercepts any `Edit` or `Write` tool call from Claude Code and **blocks it**, returning a message that instructs Claude to use the equivalent cctools command via Bash instead.

- `Edit` → `cctools edit --file <path> --old "<old>" --new "<new>"`
- `Write` → `cctools write --file <path> --content "<content>"`
- `Read` is not blocked (it doesn't modify files), but Claude is instructed to use `cctools read --detect-encoding` for encoding awareness.

## Verify

After setup, restart Claude Code and ask it to edit a file. It should use `cctools edit` via Bash instead of the built-in Edit tool.

## Matching strategies for legacy code

When working with irregular or legacy code, tell Claude to use these flags:

| Flag | Use case |
|------|----------|
| `--auto-normalize` | Tolerates whitespace differences (tabs vs spaces) |
| `--fuzzy --similarity 0.7` | Similarity-based matching |
| `--aggressive-fuzzy` | Ultra-tolerant keyword matching |
| `--smart-code` | Understands code structure (blocks, functions) |
| `--preview` | Always verify before applying |
