# CCTools - File Editing CLI with Encoding Preservation

[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Cross Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey.svg)](#downloads)

CCTools is a command-line interface for file editing operations that **automatically preserves the original encoding** of files during edits. Perfect for working with legacy codebases, international projects, and files with various character encodings.

## 🚀 Features

### File Operations
- **Automatic Encoding Detection**: Detects file encoding using advanced algorithms
- **Encoding Preservation**: Maintains original file encoding during all operations
- **Atomic Operations**: All-or-nothing approach for multiple edits
- **Safe File Operations**: Copy, move, and delete with backup support

### Directory Operations
- **Recursive Directory Operations**: Copy, move, and delete entire directory trees
- **Intelligent Directory Listing**: Analyze encoding distribution across projects
- **Atomic Directory Operations**: Complete success or full rollback
- **Backup and Recovery**: Comprehensive backup system for directories

### Cross-Platform & Safety
- **Cross-Platform**: Works on Windows, Linux, and macOS
- **Safe Operations**: Automatic backup and rollback on failures
- **Multiple Formats**: Supports UTF-8, ISO-8859-1, Windows-1252, and many more
- **Progress Tracking**: Detailed statistics and progress reporting

## 📥 Installation

### Download Pre-built Binaries

Download the latest release for your platform:

- **Windows**: Download `cctools_windows_amd64.zip` or `cctools_windows_386.zip`
- **Linux**: Download `cctools_linux_amd64.tar.gz`, `cctools_linux_386.tar.gz`, or `cctools_linux_arm64.tar.gz`
- **macOS**: Download `cctools_darwin_amd64.tar.gz` (Intel) or `cctools_darwin_arm64.tar.gz` (Apple Silicon)

### Build from Source

```bash
git clone https://github.com/your-username/cctools.git
cd cctools
go build -o cctools
```

### Cross-Platform Build

```bash
chmod +x build.sh
./build.sh
```

This will create binaries for all supported platforms in the `dist/` directory.

## 🔧 Usage

### Basic Commands

CCTools provides comprehensive file and directory operations:

```bash
cctools [command] [flags]
```

#### File Operations:
- `read` - Read files with automatic encoding detection
- `write` - Create/overwrite files with specified encoding
- `edit` - Edit files by replacing text strings
- `multiedit` - Perform multiple edit operations atomically
- `copy` - Copy files preserving encoding
- `move` - Move files with atomic rollback
- `delete` - Delete files with optional backup

#### Directory Operations:
- `mkdir` - Create directories with proper permissions
- `copydir` - Copy directories recursively preserving encodings
- `movedir` - Move directories with atomic operations
- `rmdir` - Remove directories with backup support
- `listdir` - List directory contents with encoding analysis

### 📖 Read Files

```bash
# Read file content (auto-converts to UTF-8 for display)
cctools read --file myfile.txt

# Only detect and show encoding
cctools read --file legacy.pas --detect-encoding

# Verbose output
cctools read --file config.ini --verbose
```

### ✏️ Write Files

```bash
# Create new file (UTF-8 by default)
cctools write --file newfile.txt --content "Hello World"

# Create with specific encoding
cctools write --file legacy.pas --content "unit MyUnit;" --encoding ISO-8859-1

# Overwrite existing file
cctools write --file config.ini --content "[settings]\ndebug=true" --verbose
```

### 🔄 Edit Files

```bash
# Simple replacement (preserves original encoding)
cctools edit --file config.ini --old "debug=false" --new "debug=true"

# Replace all occurrences
cctools edit --file script.js --old "console.log" --new "logger.info" --replace-all

# Verbose output
cctools edit --file main.pas --old "version := '1.0'" --new "version := '2.0'" --verbose
```

### 🎯 Multiple Edits

Create a JSON file with multiple edits:

```json
{
  "file_path": "/path/to/file.txt",
  "edits": [
    {
      "old_string": "debug: false",
      "new_string": "debug: true",
      "replace_all": false
    },
    {
      "old_string": "port: 3000",
      "new_string": "port: 8080",
      "replace_all": false
    }
  ]
}
```

Then apply all edits atomically:

```bash
cctools multiedit --edits-file changes.json --verbose
```

### 📁 File Operations

```bash
# Copy files preserving encoding
cctools copy --source arquivo.pas --dest backup.pas --preserve-mode

# Move files safely
cctools move --source old_config.ini --dest new_config.ini

# Delete with backup
cctools delete --file temp.log --backup --backup-path /safe/temp.log.bak
```

### 📂 Directory Operations

```bash
# Create directory structures
cctools mkdir --path projeto/src/main --parents

# Copy entire projects preserving all encodings
cctools copydir --source old_project/ --dest backup/ --preserve-all

# Move directories atomically
cctools movedir --source temp_project/ --dest archive/

# Remove with backup
cctools rmdir --path old_data/ --recursive --backup

# Analyze project encodings
cctools listdir --path . --recursive --show-encoding --verbose
```

## 🌍 Supported Encodings

- UTF-8 (default for new files)
- UTF-16 (LE/BE)
- ISO-8859-1
- ISO-8859-15
- Windows-1252
- Windows-1251
- GB18030, GBK
- Big5
- Shift_JIS, EUC-JP
- EUC-KR

## 💡 Why CCTools?

### The Problem
Standard file and directory tools often:
- Assume UTF-8 encoding everywhere
- Corrupt special characters in legacy files
- Don't preserve original file encoding
- Break compatibility with older systems
- Lack atomic operations for complex changes
- Don't provide adequate backup/recovery

### The Solution
CCTools:
- **Detects encoding automatically** before any operation
- **Preserves original encoding** in all file and directory operations
- **Maintains compatibility** with legacy systems
- **Provides atomic operations** for safety
- **Comprehensive backup system** for recovery
- **Intelligent directory operations** with encoding analysis

## 🔒 Safety Features

### File Operations
- **Automatic Backup**: Creates backup before editing/deleting
- **Rollback on Failure**: Restores original if operation fails
- **Atomic Multi-Edits**: All edits succeed or all are reverted
- **Overwrite Protection**: Prevents accidental file overwrites
- **Encoding Verification**: Confirms encoding before and after operations

### Directory Operations
- **Complete Directory Backup**: Full structure backup before destructive operations
- **Atomic Directory Operations**: Complete success or full rollback
- **Progressive Operations**: Skip existing files, preserve permissions
- **Intelligent Conflict Resolution**: Handle existing destinations safely
- **Comprehensive Logging**: Detailed statistics and progress tracking

## 📚 Examples

### Working with Legacy Pascal Code

```bash
# Check encoding of legacy Pascal file
cctools read --file sistema.pas --detect-encoding
# Output: Detected encoding: ISO-8859-1

# Edit while preserving ISO-8859-1 encoding
cctools edit --file sistema.pas --old "versão := '1.0'" --new "versão := '2.0'"

# Verify encoding is still preserved
cctools read --file sistema.pas --detect-encoding
# Output: Detected encoding: ISO-8859-1
```

### Batch Configuration Updates

```bash
# Create edits file
cat > config_updates.json << EOF
{
  "file_path": "app.config",
  "edits": [
    {"old_string": "debug=false", "new_string": "debug=true"},
    {"old_string": "timeout=30", "new_string": "timeout=60"},
    {"old_string": "host=localhost", "new_string": "host=0.0.0.0"}
  ]
}
EOF

# Apply all changes atomically
cctools multiedit --edits-file config_updates.json
```

### Complete Project Migration

```bash
# Analyze current project encodings
cctools listdir --path . --recursive --show-encoding --verbose

# Create backup of entire project
cctools copydir --source . --dest ../backup_$(date +%Y%m%d)/ --preserve-all

# Reorganize project structure
cctools mkdir --path new_structure/src/main --parents
cctools movedir --source old_modules/ --dest new_structure/src/

# Clean up old files safely
cctools rmdir --path temp/ --recursive --backup
```

### Encoding Analysis and Cleanup

```bash
# Analyze encoding distribution in large projects
cctools listdir --path /legacy_codebase --recursive --show-encoding > encoding_report.txt

# Copy only specific file types preserving encodings
cctools listdir --path src/ --filter "*.pas" --show-encoding
cctools copydir --source src/ --dest pascal_backup/ --preserve-all

# Safe cleanup with comprehensive backup
cctools rmdir --path old_version/ --recursive --backup --backup-path /safe/old_version_backup/
```

## 🏗️ Architecture

```
cctools/
├── cmd/                    # Cobra CLI commands
│   ├── root.go            # Root command
│   ├── read.go            # Read command
│   ├── write.go           # Write command
│   ├── edit.go            # Edit command
│   ├── multiedit.go       # Multi-edit command
│   ├── copy.go            # File copy command
│   ├── move.go            # File move command
│   ├── delete.go          # File delete command
│   ├── mkdir.go           # Directory creation command
│   ├── copydir.go         # Directory copy command
│   ├── movedir.go         # Directory move command
│   ├── rmdir.go           # Directory removal command
│   └── listdir.go         # Directory listing command
├── pkg/
│   ├── encoding/          # Encoding detection & conversion
│   └── fileops/           # File and directory operations
├── internal/models/       # Data structures
├── docs/                  # Documentation
│   ├── tools.md          # Technical documentation
│   └── prompt.md         # AI usage instructions
└── dist/                  # Built binaries (created by build.sh)
```

## 🔧 Development

### Prerequisites

- Go 1.24+
- Git

### Building

```bash
# Install dependencies
go mod tidy

# Build for current platform
go build -o cctools

# Cross-platform build
./build.sh
```

### Testing

```bash
# Test file operations with various encodings
cctools read --file test_files/utf8.txt --detect-encoding
cctools read --file test_files/iso88591.txt --detect-encoding
cctools read --file test_files/windows1252.txt --detect-encoding

# Test directory operations
cctools listdir --path test_project/ --recursive --show-encoding
cctools copydir --source test_project/ --dest backup/ --preserve-all
cctools movedir --source backup/ --dest archive/

# Test comprehensive workflows
cctools mkdir --path testing/structure --parents
cctools copy --source config.ini --dest testing/config_backup.ini
cctools delete --file testing/temp.log --backup
```

## 🙏 Acknowledgments

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [chardet](https://github.com/saintfish/chardet) - Character encoding detection
- [golang.org/x/text](https://golang.org/x/text) - Text processing

---

## 🏆 Recent Updates

### Version 1.0.0 - Complete File & Directory Operations
- ✅ **New File Operations**: `copy`, `move`, `delete` with encoding preservation
- ✅ **New Directory Operations**: `mkdir`, `copydir`, `movedir`, `rmdir`, `listdir`
- ✅ **Enhanced Safety**: Comprehensive backup and rollback systems
- ✅ **Intelligent Analysis**: Encoding distribution analysis for entire projects
- ✅ **Atomic Operations**: Complete success or full rollback for all operations
- ✅ **Cross-Platform**: Full compatibility with Windows, Linux, and macOS
- ✅ **100% Tested**: All operations tested with comprehensive test suite

---