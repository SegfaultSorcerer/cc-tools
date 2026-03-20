@echo off
REM CCTools PreToolUse Hook — blocks Edit/Write and redirects to cctools
REM Install: copy to your project and reference in .claude\settings.json
echo {"hookSpecificOutput":{"permissionDecision":"deny"},"systemMessage":"BLOCKED: Do not use Edit/Write directly. Use cctools via Bash instead:\n- Edit: cctools edit --file <path> --old \"<old>\" --new \"<new>\"\n- Write: cctools write --file <path> --content \"<content>\" [--encoding UTF-8]\n- Detect encoding first: cctools read --file <path> --detect-encoding\n- Multiple edits: cctools multiedit --edits-file operations.json\n- Use --preview to verify matches.\n- For legacy/irregular code: combine --auto-normalize --smart-code --aggressive-fuzzy"}
