package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"cctools/pkg/fileops"

	"github.com/spf13/cobra"
)

var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "Write content to a file with specified encoding",
	Long: `Write content to a file with the specified encoding.

If the file already exists, it will be overwritten completely.
If no encoding is specified, UTF-8 is used by default.

Content can be provided in three ways (exactly one must be used):
- --content: Specify content directly as a command line argument
- --content-file: Read content from another file
- --stdin: Read content from standard input (pipe or interactive)

Examples:
  # Direct content
  cctools write --file output.txt --content "Hello World"

  # From file
  cctools write --file backup.txt --content-file original.txt

  # From stdin (pipe)
  echo "Hello World" | cctools write --file output.txt --stdin

  # From stdin (interactive)
  cctools write --file output.txt --stdin

  # With specific encoding
  cctools write --file output.txt --content "Hello" --encoding ISO-8859-1`,
	RunE: runWriteCmd,
}

var (
	writeFilePath    string
	writeContent     string
	writeContentFile string
	writeFromStdin   bool
	writeEncoding    string
)

func init() {
	rootCmd.AddCommand(writeCmd)

	writeCmd.Flags().StringVarP(&writeFilePath, "file", "f", "", "Path to the file to write (required)")
	writeCmd.Flags().StringVarP(&writeContent, "content", "c", "", "Content to write to the file")
	writeCmd.Flags().StringVar(&writeContentFile, "content-file", "", "Read content from specified file")
	writeCmd.Flags().BoolVar(&writeFromStdin, "stdin", false, "Read content from stdin")
	writeCmd.Flags().StringVarP(&writeEncoding, "encoding", "e", "UTF-8", "Encoding to use for the file (default: UTF-8)")

	writeCmd.MarkFlagRequired("file")
}

func runWriteCmd(cmd *cobra.Command, args []string) error {
	// Convert to absolute path if relative
	absPath, err := filepath.Abs(writeFilePath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Determine content source and get content
	content, err := getContentFromSource()
	if err != nil {
		return fmt.Errorf("failed to get content: %w", err)
	}

	// Process unicode escapes when content comes from --content flag
	if writeContent != "" {
		content = processWriteUnicodeEscapes(content)
	}

	// Create file operations handler
	fileOps := fileops.NewFileOperations()

	// Write the file
	if err := fileOps.WriteFile(absPath, content, writeEncoding); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Get verbose flag from parent command
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		fmt.Printf("Successfully wrote file: %s\n", absPath)
		fmt.Printf("Encoding: %s\n", writeEncoding)
		fmt.Printf("Content length: %d characters\n", len(content))
		if writeFromStdin {
			fmt.Printf("Content source: stdin\n")
		} else if writeContentFile != "" {
			fmt.Printf("Content source: %s\n", writeContentFile)
		} else {
			fmt.Printf("Content source: command line\n")
		}
	} else {
		fmt.Printf("File written successfully: %s\n", absPath)
	}

	return nil
}

// getContentFromSource determines the content source and retrieves content
func getContentFromSource() (string, error) {
	sources := 0
	if writeContent != "" {
		sources++
	}
	if writeContentFile != "" {
		sources++
	}
	if writeFromStdin {
		sources++
	}

	// Exactly one source must be specified
	if sources == 0 {
		return "", fmt.Errorf("no content source specified: use --content, --content-file, or --stdin")
	}
	if sources > 1 {
		return "", fmt.Errorf("multiple content sources specified: use only one of --content, --content-file, or --stdin")
	}

	// Get content from the specified source
	if writeFromStdin {
		return readFromStdin()
	}
	if writeContentFile != "" {
		return readFromFile(writeContentFile)
	}
	return writeContent, nil
}

// readFromStdin reads content from standard input
func readFromStdin() (string, error) {
	// Check if stdin has data available
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to check stdin: %w", err)
	}

	// If stdin is a terminal (no pipe/redirect), prompt user
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		fmt.Println("Reading from stdin. Type your content and press Ctrl+D (Unix) or Ctrl+Z (Windows) when done:")
	}

	// Read all content from stdin
	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read from stdin: %w", err)
	}

	return string(content), nil
}

// readFromFile reads content from a specified file
func readFromFile(filePath string) (string, error) {
	// Convert to absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", fmt.Errorf("content file does not exist: %s", absPath)
	}

	// Read file content
	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read content file: %w", err)
	}

	return string(content), nil
}

// processWriteUnicodeEscapes converts \uXXXX sequences to actual unicode characters
func processWriteUnicodeEscapes(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if i+5 < len(s) && s[i] == '\\' && s[i+1] == 'u' {
			hex := s[i+2 : i+6]
			if codepoint, err := strconv.ParseInt(hex, 16, 32); err == nil {
				result.WriteRune(rune(codepoint))
				i += 6
				continue
			}
		}
		result.WriteByte(s[i])
		i++
	}
	return result.String()
}