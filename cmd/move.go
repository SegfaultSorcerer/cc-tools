package cmd

import (
	"fmt"
	"path/filepath"

	"cctools/internal/models"
	"cctools/pkg/fileops"

	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:   "move",
	Short: "Move a file to another location preserving encoding",
	Long: `Move a file from source to destination preserving the original encoding.

The command automatically detects the source file encoding and moves it
exactly as-is to the destination path. This ensures no character corruption
occurs during the move operation. The operation is atomic - if any step fails,
the source file is restored from backup.

Examples:
  cctools move --source file.txt --dest /backup/file.txt
  cctools move --source sistema.pas --dest /new/location/sistema.pas
  cctools move -s old_config.ini -d new_config.ini --overwrite`,
	RunE: runMoveCmd,
}

var (
	moveSourcePath string
	moveDestPath   string
	moveOverwrite  bool
)

func init() {
	rootCmd.AddCommand(moveCmd)

	moveCmd.Flags().StringVarP(&moveSourcePath, "source", "s", "", "Source file path (required)")
	moveCmd.Flags().StringVarP(&moveDestPath, "dest", "d", "", "Destination file path (required)")
	moveCmd.Flags().BoolVarP(&moveOverwrite, "overwrite", "o", false, "Overwrite destination if it exists")

	moveCmd.MarkFlagRequired("source")
	moveCmd.MarkFlagRequired("dest")
}

func runMoveCmd(cmd *cobra.Command, args []string) error {
	// Convert to absolute paths
	absSourcePath, err := filepath.Abs(moveSourcePath)
	if err != nil {
		return fmt.Errorf("failed to resolve source path: %w", err)
	}

	absDestPath, err := filepath.Abs(moveDestPath)
	if err != nil {
		return fmt.Errorf("failed to resolve destination path: %w", err)
	}

	// Create file operations handler
	fileOps := fileops.NewFileOperations()

	// Get verbose flag from parent command
	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		fmt.Printf("Moving file from %s to %s\n", absSourcePath, absDestPath)
		if moveOverwrite {
			fmt.Println("Will overwrite destination if it exists")
		}
	}

	// Create move operation
	operation := &models.MoveOperation{
		SourcePath:      absSourcePath,
		DestinationPath: absDestPath,
		Overwrite:       moveOverwrite,
	}

	// Execute move operation
	result, err := fileOps.MoveFile(operation)
	if err != nil {
		return fmt.Errorf("move failed: %s", result.Message)
	}

	// Display results
	if verbose {
		fmt.Printf("Original encoding: %s\n", result.SourceInfo.Encoding)
		fmt.Printf("File size: %d bytes\n", len(result.SourceInfo.Content))
		fmt.Printf("Final encoding: %s\n", result.TargetInfo.Encoding)
	}

	fmt.Println(result.Message)
	return nil
}