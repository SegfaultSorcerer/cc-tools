package cmd

import (
	"fmt"
	"path/filepath"

	"cctools/internal/models"
	"cctools/pkg/fileops"

	"github.com/spf13/cobra"
)

var movedirCmd = &cobra.Command{
	Use:   "movedir",
	Short: "Move directories preserving encoding with rollback support",
	Long: `Move directories and their contents while preserving the original encoding
of all files. The operation is atomic - if any step fails, the source
directory is restored to its original state.

The command tries to use efficient rename operation first, falling back
to copy + delete with full rollback support if needed.

Examples:
  cctools movedir --source oldproject/ --dest newlocation/
  cctools movedir --source temp/ --dest archive/temp/ --overwrite
  cctools movedir -s project/v1/ -d project/v2/ --overwrite`,
	RunE: runMovedirCmd,
}

var (
	movedirSourcePath string
	movedirDestPath   string
	movedirOverwrite  bool
)

func init() {
	rootCmd.AddCommand(movedirCmd)

	movedirCmd.Flags().StringVarP(&movedirSourcePath, "source", "s", "", "Source directory path (required)")
	movedirCmd.Flags().StringVarP(&movedirDestPath, "dest", "d", "", "Destination directory path (required)")
	movedirCmd.Flags().BoolVarP(&movedirOverwrite, "overwrite", "o", false, "Overwrite destination if it exists")

	movedirCmd.MarkFlagRequired("source")
	movedirCmd.MarkFlagRequired("dest")
}

func runMovedirCmd(cmd *cobra.Command, args []string) error {
	// Convert to absolute paths
	absSourcePath, err := filepath.Abs(movedirSourcePath)
	if err != nil {
		return fmt.Errorf("failed to resolve source path: %w", err)
	}

	absDestPath, err := filepath.Abs(movedirDestPath)
	if err != nil {
		return fmt.Errorf("failed to resolve destination path: %w", err)
	}

	// Create file operations handler
	fileOps := fileops.NewFileOperations()

	// Get verbose flag from parent command
	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		fmt.Printf("Moving directory from %s to %s\n", absSourcePath, absDestPath)
		if movedirOverwrite {
			fmt.Println("Will overwrite destination if it exists")
		}
	}

	// Create move operation
	operation := &models.DirectoryMoveOperation{
		SourcePath:      absSourcePath,
		DestinationPath: absDestPath,
		Overwrite:       movedirOverwrite,
	}

	// Execute move operation
	result, err := fileOps.MoveDirectory(operation)
	if err != nil {
		return fmt.Errorf("movedir failed: %s", result.Message)
	}

	// Display results
	if verbose {
		if result.ProcessedFiles > 0 {
			fmt.Printf("Files processed: %d\n", result.ProcessedFiles)
			fmt.Printf("Directories processed: %d\n", result.ProcessedDirs)
			fmt.Printf("Total size: %d bytes\n", result.TotalSize)
		}

		if result.SourceInfo != nil && len(result.SourceInfo.Encodings) > 0 {
			fmt.Println("Encodings preserved:")
			for encoding, count := range result.SourceInfo.Encodings {
				fmt.Printf("  %s: %d files\n", encoding, count)
			}
		}
	}

	fmt.Println(result.Message)
	return nil
}