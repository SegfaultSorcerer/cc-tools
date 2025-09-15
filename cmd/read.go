package cmd

import (
	"fmt"
	"path/filepath"

	"cctools/pkg/encoding"
	"cctools/pkg/fileops"

	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Read a file with automatic encoding detection",
	Long: `Read a file and display its content with automatic encoding detection.

The command detects the file encoding automatically and converts the content
to UTF-8 for display. Use --detect-encoding to show only the detected encoding
without displaying the file content.

Examples:
  cctools read --file /path/to/file.txt
  cctools read --file file.txt --detect-encoding
  cctools read -f file.txt -d`,
	RunE: runReadCmd,
}

var (
	readFilePath      string
	readDetectOnly    bool
)

func init() {
	rootCmd.AddCommand(readCmd)

	readCmd.Flags().StringVarP(&readFilePath, "file", "f", "", "Path to the file to read (required)")
	readCmd.Flags().BoolVarP(&readDetectOnly, "detect-encoding", "d", false, "Only detect and show the file encoding")

	readCmd.MarkFlagRequired("file")
}

func runReadCmd(cmd *cobra.Command, args []string) error {
	// Convert to absolute path if relative
	absPath, err := filepath.Abs(readFilePath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Create file operations handler
	fileOps := fileops.NewFileOperations()

	// Read the file
	fileInfo, err := fileOps.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Get verbose flag from parent command
	verbose, _ := cmd.Flags().GetBool("verbose")

	if readDetectOnly {
		// Only show encoding information
		fmt.Printf("File: %s\n", absPath)
		fmt.Printf("Detected encoding: %s\n", fileInfo.Encoding)
		if verbose {
			fmt.Printf("File size: %d bytes\n", len(fileInfo.Content))
		}
		return nil
	}

	// Decode and display content
	content, err := encoding.DecodeBytes(fileInfo.Content, fileInfo.Encoding)
	if err != nil {
		return fmt.Errorf("failed to decode file content: %w", err)
	}

	if verbose {
		fmt.Printf("File: %s\n", absPath)
		fmt.Printf("Detected encoding: %s\n", fileInfo.Encoding)
		fmt.Printf("Content length: %d characters\n", len(content))
		fmt.Printf("File size: %d bytes\n", len(fileInfo.Content))
		fmt.Println("--- Content ---")
	}

	fmt.Print(content)

	// Add newline if file doesn't end with one and we're in verbose mode
	if verbose && len(content) > 0 && content[len(content)-1] != '\n' {
		fmt.Println()
		fmt.Println("--- End of file ---")
	}

	return nil
}