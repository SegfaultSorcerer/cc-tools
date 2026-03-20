#!/bin/bash
# CCTools PreToolUse Hook — blocks Edit/Write and redirects to cctools
# Install: copy to your project and reference in .claude/settings.json

cat <<'EOF'
{
  "hookSpecificOutput": {
    "permissionDecision": "deny"
  },
  "systemMessage": "BLOCKED: Do not use the Edit/Write tool directly. This project uses cctools for file operations to preserve encoding. Use these Bash commands instead:\n- Edit: cctools edit --file <path> --old \"<old>\" --new \"<new>\"\n- Write: cctools write --file <path> --content \"<content>\" [--encoding UTF-8]\n- Always detect encoding first: cctools read --file <path> --detect-encoding\n- For multiple edits: cctools multiedit --edits-file operations.json\n- Use --preview flag to verify matches before applying.\n- For legacy/irregular code: combine --auto-normalize --smart-code --aggressive-fuzzy"
}
EOF
