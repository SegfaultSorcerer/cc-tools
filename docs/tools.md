# File Editing Tools

This document describes the tools available for file editing in Claude Code, including their parameters, operation, and technical characteristics.

## Overview

Claude Code has 4 main tools for file editing:

1. **Edit** - Simple editing with text replacement
2. **MultiEdit** - Multiple edits in a single file
3. **Write** - Complete creation/overwriting of files
4. **NotebookEdit** - Specific editing for Jupyter notebooks

## 1. Edit Tool

### Description
Performs exact string replacements in existing files.

### Parameters
- `file_path` (required): Absolute path of the file to be modified
- `old_string` (required): Text to be replaced
- `new_string` (required): Replacement text
- `replace_all` (optional, default: false): Replaces all occurrences

### How It Works
- Searches for the exact string in `old_string` in the file
- Replaces with `new_string`
- The string must be unique in the file (except if `replace_all=true`)
- Preserves exactly the original indentation and formatting

### Requirements
- The file must be read with the `Read` tool before editing
- Paths must be absolute (start with `/`)
- The `old_string` must exist exactly as specified

### Example
```json
{
  "file_path": "/home/user/project/app.js",
  "old_string": "const port = 3000;",
  "new_string": "const port = process.env.PORT || 3000;"
}
```

## 2. MultiEdit Tool

### Description
Allows multiple edits in a single file atomically.

### Parameters
- `file_path` (required): Absolute path of the file
- `edits` (required): Array of edit objects, each containing:
  - `old_string`: Text to be replaced
  - `new_string`: Replacement text
  - `replace_all` (optional): Replaces all occurrences

### How It Works
- Applies all edits sequentially
- If any edit fails, none are applied (atomic operation)
- Each edit operates on the result of the previous edit
- Ideal for multiple changes in the same file

### Requirements
- File must be read previously
- All edits must be valid for the operation to succeed
- Plan carefully to avoid conflicts between sequential edits

### Example
```json
{
  "file_path": "/home/user/project/config.js",
  "edits": [
    {
      "old_string": "debug: false",
      "new_string": "debug: true"
    },
    {
      "old_string": "port: 3000",
      "new_string": "port: 8080"
    }
  ]
}
```

## 3. Write Tool

### Description
Creates new files or completely overwrites existing files.

### Parameters
- `file_path` (required): Absolute path of the file
- `content` (required): Complete content of the file

### How It Works
- Completely overwrites the file if it exists
- Creates new file if it doesn't exist
- Replaces all previous content

### Requirements
- For existing files, must be read previously
- Paths must be absolute
- Prefer editing existing files over creation

### Example
```json
{
  "file_path": "/home/user/project/new-file.js",
  "content": "console.log('New file created!');\nmodule.exports = {};"
}
```

## 4. NotebookEdit Tool

### Description
Specific editing for Jupyter Notebook files (.ipynb).

### Parameters
- `notebook_path` (required): Absolute path of the notebook
- `new_source` (required): New content of the cell
- `cell_id` (optional): ID of the cell to be edited
- `cell_type` (optional): Type of cell ("code" or "markdown")
- `edit_mode` (optional): Edit mode ("replace", "insert", "delete")

### How It Works
- `replace`: Replaces content of existing cell
- `insert`: Adds new cell
- `delete`: Removes existing cell
- Maintains JSON structure of the notebook

## File Encoding

### Default Encoding
- **UTF-8**: All files are saved in UTF-8 by default
- Full support for Unicode characters
- Compatible with special characters, accents, and emojis

### Persistence Behavior
- Files are saved immediately after each operation
- No cache or temporary buffer
- Changes are persisted directly to the file system
- Preserves existing file permissions

### Technical Characteristics
- Line endings preserved according to operating system
- Indentation (tabs/spaces) preserved exactly
- Control characters maintained when present in original

## Best Practices

### Before Editing
1. Always use `Read` to examine the file first
2. Check existing code conventions
3. Understand the project structure

### During Editing
1. Always use absolute paths
2. Preserve original indentation and formatting
3. Test complex edits with `MultiEdit` when appropriate

### After Editing
1. Verify that changes were applied correctly
2. Run tests when available
3. Check lint/typecheck if configured in the project

## Important Limitations

- **Strings must be exact**: Spaces and indentation must match perfectly
- **Absolute paths required**: Relative paths are not accepted
- **Previous reading mandatory**: Existing files must be read before editing
- **Atomic operations**: MultiEdit fails completely if any individual edit fails
- **No undo**: No integrated undo functionality