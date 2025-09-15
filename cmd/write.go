package cmd

import (
	"fmt"
	"path/filepath"

	"cctools/pkg/fileops"

	"github.com/spf13/cobra"
)

var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "Write content to a file with specified encoding",
	Long: `Write content to a file with the specified encoding.

If the file already exists, it will be overwritten completely.
If no encoding is specified, UTF-8 is used by default.

Examples:
  cctools write --file /path/to/file.txt --content "Hello World"
  cctools write --file /path/to/file.txt --content "Hello World" --encoding ISO-8859-1
  cctools write -f file.txt -c "Content" -e UTF-8`,
	RunE: runWriteCmd,
}

var (
	writeFilePath string
	writeContent  string
	writeEncoding string
)

func init() {
	rootCmd.AddCommand(writeCmd)

	writeCmd.Flags().StringVarP(&writeFilePath, "file", "f", "", "Path to the file to write (required)")
	writeCmd.Flags().StringVarP(&writeContent, "content", "c", "", "Content to write to the file (required)")
	writeCmd.Flags().StringVarP(&writeEncoding, "encoding", "e", "UTF-8", "Encoding to use for the file (default: UTF-8)")

	writeCmd.MarkFlagRequired("file")
	writeCmd.MarkFlagRequired("content")
}

func runWriteCmd(cmd *cobra.Command, args []string) error {
	// Convert to absolute path if relative
	absPath, err := filepath.Abs(writeFilePath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Create file operations handler
	fileOps := fileops.NewFileOperations()

	// Write the file
	if err := fileOps.WriteFile(absPath, writeContent, writeEncoding); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Get verbose flag from parent command
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		fmt.Printf("Successfully wrote file: %s\n", absPath)
		fmt.Printf("Encoding: %s\n", writeEncoding)
		fmt.Printf("Content length: %d characters\n", len(writeContent))
	} else {
		fmt.Printf("File written successfully: %s\n", absPath)
	}

	return nil
}