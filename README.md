# CCTools - File Editing CLI with Encoding Preservation

[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Cross Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey.svg)](#downloads)

CCTools is a command-line interface for file editing operations that **automatically preserves the original encoding** of files during edits. Perfect for working with legacy codebases, international projects, and files with various character encodings.

## 🚀 Features

- **Automatic Encoding Detection**: Detects file encoding using advanced algorithms
- **Encoding Preservation**: Maintains original file encoding during edits
- **Atomic Operations**: All-or-nothing approach for multiple edits
- **Cross-Platform**: Works on Windows, Linux, and macOS
- **Safe Operations**: Automatic backup and rollback on failures
- **Multiple Formats**: Supports UTF-8, ISO-8859-1, Windows-1252, and many more

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

CCTools provides four main commands:

```bash
cctools [command] [flags]
```

#### Available Commands:
- `read` - Read files with automatic encoding detection
- `write` - Create/overwrite files with specified encoding
- `edit` - Edit files by replacing text strings
- `multiedit` - Perform multiple edit operations atomically

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
Standard file editing tools often:
- Assume UTF-8 encoding
- Corrupt special characters in legacy files
- Don't preserve original file encoding
- Break compatibility with older systems

### The Solution
CCTools:
- **Detects encoding automatically** before any operation
- **Preserves original encoding** during edits
- **Maintains compatibility** with legacy systems
- **Provides atomic operations** for safety

## 🔒 Safety Features

- **Automatic Backup**: Creates backup before editing
- **Rollback on Failure**: Restores original if operation fails
- **Atomic Multi-Edits**: All edits succeed or all are reverted
- **Validation**: Checks for unique strings and validates operations
- **Encoding Verification**: Confirms encoding before and after operations

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

## 🏗️ Architecture

```
cctools/
├── cmd/                    # Cobra CLI commands
│   ├── root.go            # Root command
│   ├── read.go            # Read command
│   ├── write.go           # Write command
│   ├── edit.go            # Edit command
│   └── multiedit.go       # Multi-edit command
├── pkg/
│   ├── encoding/          # Encoding detection & conversion
│   └── fileops/           # File operations
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
# Test with various file encodings
cctools read --file test_files/utf8.txt --detect-encoding
cctools read --file test_files/iso88591.txt --detect-encoding
cctools read --file test_files/windows1252.txt --detect-encoding
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: Check the `docs/` directory
- **Issues**: Report bugs on GitHub Issues
- **Discussions**: Use GitHub Discussions for questions

## 🙏 Acknowledgments

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [chardet](https://github.com/saintfish/chardet) - Character encoding detection
- [golang.org/x/text](https://golang.org/x/text) - Text processing

---

**Made with ❤️ for developers working with multi-encoding projects**