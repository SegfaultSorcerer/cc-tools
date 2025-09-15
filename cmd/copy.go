package cmd

import (
	"fmt"
	"path/filepath"

	"cctools/internal/models"
	"cctools/pkg/fileops"

	"github.com/spf13/cobra"
)

var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy a file to another location preserving encoding",
	Long: `Copy a file from source to destination preserving the original encoding.

The command automatically detects the source file encoding and copies it
exactly as-is to the destination path. This ensures no character corruption
occurs during the copy operation.

Examples:
  cctools copy --source file.txt --dest backup.txt
  cctools copy --source sistema.pas --dest /backup/sistema.pas --preserve-mode
  cctools copy -s config.ini -d /new/path/config.ini --overwrite`,
	RunE: runCopyCmd,
}

var (
	copySourcePath      string
	copyDestPath        string
	copyPreserveMode    bool
	copyOverwrite       bool
)

func init() {
	rootCmd.AddCommand(copyCmd)

	copyCmd.Flags().StringVarP(&copySourcePath, "source", "s", "", "Source file path (required)")
	copyCmd.Flags().StringVarP(&copyDestPath, "dest", "d", "", "Destination file path (required)")
	copyCmd.Flags().BoolVarP(&copyPreserveMode, "preserve-mode", "p", false, "Preserve file permissions")
	copyCmd.Flags().BoolVarP(&copyOverwrite, "overwrite", "o", false, "Overwrite destination if it exists")

	copyCmd.MarkFlagRequired("source")
	copyCmd.MarkFlagRequired("dest")
}

func runCopyCmd(cmd *cobra.Command, args []string) error {
	// Convert to absolute paths
	absSourcePath, err := filepath.Abs(copySourcePath)
	if err != nil {
		return fmt.Errorf("failed to resolve source path: %w", err)
	}

	absDestPath, err := filepath.Abs(copyDestPath)
	if err != nil {
		return fmt.Errorf("failed to resolve destination path: %w", err)
	}

	// Create file operations handler
	fileOps := fileops.NewFileOperations()

	// Get verbose flag from parent command
	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		fmt.Printf("Copying file from %s to %s\n", absSourcePath, absDestPath)
		if copyPreserveMode {
			fmt.Println("File permissions will be preserved")
		}
		if copyOverwrite {
			fmt.Println("Will overwrite destination if it exists")
		}
	}

	// Create copy operation
	operation := &models.CopyOperation{
		SourcePath:      absSourcePath,
		DestinationPath: absDestPath,
		PreserveMode:    copyPreserveMode,
		Overwrite:       copyOverwrite,
	}

	// Execute copy operation
	result, err := fileOps.CopyFile(operation)
	if err != nil {
		return fmt.Errorf("copy failed: %s", result.Message)
	}

	// Display results
	if verbose {
		fmt.Printf("Source encoding: %s\n", result.SourceInfo.Encoding)
		fmt.Printf("Destination encoding: %s\n", result.TargetInfo.Encoding)
		fmt.Printf("File size: %d bytes\n", len(result.SourceInfo.Content))
	}

	fmt.Println(result.Message)
	return nil
}