# CCTools - File Editing Tool with Encoding Preservation

## Context

Standard file writing tools must be **replaced** by CCTools tools to ensure proper encoding preservation. ALWAYS use CCTools commands instead of standard editing tools.

## Base Command

```bash
cctools [command] [flags]
```

---

## MAIN COMMANDS

### 1. READ - File Reading
```bash
cctools read --file <path> [--detect-encoding] [--verbose]
```
- **Always execute before editing** to detect encoding
- Automatically detects and converts to UTF-8

### 2. WRITE - Creation/Overwrite
```bash
cctools write --file <path> --content <content> [--encoding <encoding>]
cctools write --file <path> --content-file <source_file>
cctools write --file <path> --stdin
```

### 3. EDIT - Edit by Replacement ⭐
```bash
cctools edit --file <path> --old <old_text> --new <new_text> [flags]
```

#### Essential Flags
- `--replace-all`: Replaces all occurrences
- `--preview`: Shows detailed preview without applying
- `--auto-normalize`: **[NEW]** Tolerates irregular whitespace
- `--fuzzy --similarity 0.6`: **[NEW]** Tolerant matching
- `--aggressive-fuzzy`: **[NEW]** Ultra-tolerant matching
- `--smart-code`: **[NEW]** Understands code structure
- `--smart-suggestions`: **[NEW]** Intelligent suggestions
- `--auto-chunk`: **[NEW]** For large strings

### 4. MULTIEDIT - Multiple Atomic Edits
```bash
cctools multiedit --edits-file <json_file> [--preview] [--continue-on-error]
```

---

## 🚀 CRITICAL IMPROVEMENTS 2024

### ✅ Problems Solved

#### 1. **Too Restrictive Matching** → **SOLVED**
```bash
# Before: CONSTANTLY FAILED
cctools edit -f file.pas -o "procedure   Method" -n "procedure NewMethod"

# Now: WORKS with --auto-normalize
cctools edit -f file.pas -o "procedure   Method" -n "procedure NewMethod" --auto-normalize
```

#### 2. **Large Blocks** → **SOLVED**
```bash
# Now works with --auto-chunk
cctools edit -f file.pas -o "procedure ExtensiveMethod..." -n "procedure NewMethod..." --auto-chunk
```

#### 3. **Limited Fuzzy** → **IMPROVED**
```bash
# Configurable threshold + aggressive mode
cctools edit -f file.pas -o "TaskDlg warning" -n "ShowMessage('alert')" --aggressive-fuzzy --similarity 0.3
```

#### 4. **Smart Code for Delphi/Pascal** → **NEW**
```bash
# Understands Pascal/Delphi structure
cctools edit -f file.pas -o "complex procedure" -n "new procedure" --smart-code --code-language pascal
```

#### 5. **Ambiguous Matching** → **SOLVED** ⭐
```bash
# Repetitive code in multiple functions (now intelligent selection)
cctools edit -f file.pas -o "listadez := TStringList.Create" -n "// OPTIMIZED" --debug-mode
# System calculates uniqueness and chooses most relevant match
```

### 🎯 Matching Strategies (Hierarchical)

1. **Exact Match** (default)
2. **Regex Match** (`--regex`)
3. **Smart Code** (`--smart-code`)
4. **Auto-Chunk** (`--auto-chunk`)
5. **Normalized** (`--auto-normalize`)
6. **Enhanced Fuzzy** (`--fuzzy + --similarity`)
7. **Aggressive Fuzzy** (`--aggressive-fuzzy`)

---

## DEFINITIVE COMMAND FOR DELPHI

```bash
# Ultra-tolerant for legacy code
cctools edit -f file.pas \
  --old "problematic code" \
  --new "new code" \
  --smart-code \
  --auto-normalize \
  --aggressive-fuzzy \
  --similarity 0.4 \
  --smart-suggestions \
  --preview
```

---

## IMPROVED TROUBLESHOOTING

### Problem: "String not found"
**Solutions in order of effectiveness:**
```bash
# 1. Auto-normalize (resolves 85% of cases)
cctools edit --file file.pas --old "problem text" --new "new text" --auto-normalize

# 2. Aggressive fuzzy (low threshold)
cctools edit --file file.pas --old "problem text" --new "new text" --aggressive-fuzzy --similarity 0.3

# 3. Smart code (for structures)
cctools edit --file file.pas --old "procedure structure" --new "new procedure" --smart-code

# 4. Total combination
cctools edit --file file.pas --old "problem text" --new "new text" --auto-normalize --aggressive-fuzzy --smart-suggestions
```

---

## MANDATORY RULES FOR AI

1. **NEVER use standard editing tools**
2. **ALWAYS use CCTools** for any file operations
3. **ALWAYS execute `read --detect-encoding`** before editing
4. **PREFER `edit`** over rewriting files
5. **USE `--preview` first** for complex operations
6. **For Delphi/Pascal code**: Use `--smart-code --auto-normalize --aggressive-fuzzy`

### Recommended Flow
```bash
# 1. Detect encoding
cctools read --file file.pas --detect-encoding

# 2. Preview to verify
cctools edit --file file.pas --old "old" --new "new" --preview --auto-normalize

# 3. Apply if OK
cctools edit --file file.pas --old "old" --new "new" --auto-normalize
```

### Use Cases

| Scenario | Command | Reason |
|----------|---------|--------|
| New file | `write` | Encoding control |
| Simple edit | `edit` | Preserves encoding |
| Legacy code | `edit --auto-normalize --smart-code` | Maximum tolerance |
| Multiple changes | `multiedit` | Atomic operation |
| Large strings | `edit --auto-chunk` | Breaks automatically |

---

## SUPPORTED ENCODINGS
- UTF-8 (default)
- ISO-8859-1 (common in legacy projects)
- Windows-1252, Windows-1251
- GB18030, GBK, Big5
- Shift_JIS, EUC-JP, EUC-KR

---

## EXECUTIVE SUMMARY

**CCTools v2024** ensures encoding preservation in ALL file operations with revolutionary improvements:

### 🔥 **HIGHLIGHTS:**
- ✅ **Ultra-Tolerant Matching**: `--auto-normalize` resolves 85% of "string not found"
- ✅ **Unlimited Strings**: `--auto-chunk` processes any size
- ✅ **Configurable Fuzzy**: `--similarity` 0.0-1.0 + `--aggressive-fuzzy`
- ✅ **Smart Code**: Understands Pascal/Delphi, JavaScript, Python, Go...
- ✅ **Professional Preview**: Detailed analysis with security checks

### 🎯 **CRITICAL CASES SOLVED:**
```bash
# Irregular whitespace (before always failed)
cctools edit -f file.pas -o "procedure   Method( param )" -n "procedure NewMethod(param)" --auto-normalize

# Large methods (before limited)
cctools edit -f file.pas -o "procedure ExtensiveMethod..." -n "procedure NewMethod..." --auto-chunk

# Configurable similarity (before fixed)
cctools edit -f file.txt -o "similar text" -n "new text" --aggressive-fuzzy --similarity 0.3
```

**ALWAYS use CCTools to avoid character corruption and ensure compatibility with legacy systems. The 2024 improvements make the tool practically infallible for complex string matching.**