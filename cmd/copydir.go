package cmd

import (
	"fmt"
	"path/filepath"

	"cctools/internal/models"
	"cctools/pkg/fileops"

	"github.com/spf13/cobra"
)

var copydirCmd = &cobra.Command{
	Use:   "copydir",
	Short: "Copy directories recursively preserving encoding",
	Long: `Copy directories and their contents recursively while preserving the original
encoding of all files. This ensures no character corruption occurs during
directory copy operations.

The command can preserve file permissions, timestamps, and skip existing files
for incremental copying.

Examples:
  cctools copydir --source myproject/ --dest backup/
  cctools copydir --source src/ --dest /backup/src/ --preserve-all
  cctools copydir -s olddir/ -d newdir/ --overwrite --skip-existing`,
	RunE: runCopydirCmd,
}

var (
	copydirSourcePath   string
	copydirDestPath     string
	copydirPreserveAll  bool
	copydirOverwrite    bool
	copydirSkipExisting bool
)

func init() {
	rootCmd.AddCommand(copydirCmd)

	copydirCmd.Flags().StringVarP(&copydirSourcePath, "source", "s", "", "Source directory path (required)")
	copydirCmd.Flags().StringVarP(&copydirDestPath, "dest", "d", "", "Destination directory path (required)")
	copydirCmd.Flags().BoolVar(&copydirPreserveAll, "preserve-all", false, "Preserve permissions, timestamps, and other attributes")
	copydirCmd.Flags().BoolVarP(&copydirOverwrite, "overwrite", "o", false, "Overwrite destination if it exists")
	copydirCmd.Flags().BoolVar(&copydirSkipExisting, "skip-existing", false, "Skip files that already exist in destination")

	copydirCmd.MarkFlagRequired("source")
	copydirCmd.MarkFlagRequired("dest")
}

func runCopydirCmd(cmd *cobra.Command, args []string) error {
	// Convert to absolute paths
	absSourcePath, err := filepath.Abs(copydirSourcePath)
	if err != nil {
		return fmt.Errorf("failed to resolve source path: %w", err)
	}

	absDestPath, err := filepath.Abs(copydirDestPath)
	if err != nil {
		return fmt.Errorf("failed to resolve destination path: %w", err)
	}

	// Create file operations handler
	fileOps := fileops.NewFileOperations()

	// Get verbose flag from parent command
	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		fmt.Printf("Copying directory from %s to %s\n", absSourcePath, absDestPath)
		if copydirPreserveAll {
			fmt.Println("Will preserve all file attributes")
		}
		if copydirOverwrite {
			fmt.Println("Will overwrite destination if it exists")
		}
		if copydirSkipExisting {
			fmt.Println("Will skip existing files")
		}
	}

	// Create copy operation
	operation := &models.DirectoryCopyOperation{
		SourcePath:      absSourcePath,
		DestinationPath: absDestPath,
		PreserveAll:     copydirPreserveAll,
		Overwrite:       copydirOverwrite,
		SkipExisting:    copydirSkipExisting,
	}

	// Execute copy operation
	result, err := fileOps.CopyDirectory(operation)
	if err != nil {
		return fmt.Errorf("copydir failed: %s", result.Message)
	}

	// Display results
	if verbose {
		fmt.Printf("Files processed: %d\n", result.ProcessedFiles)
		fmt.Printf("Directories processed: %d\n", result.ProcessedDirs)
		fmt.Printf("Total size: %d bytes\n", result.TotalSize)

		if result.SourceInfo != nil && len(result.SourceInfo.Encodings) > 0 {
			fmt.Println("Encodings found:")
			for encoding, count := range result.SourceInfo.Encodings {
				fmt.Printf("  %s: %d files\n", encoding, count)
			}
		}
	}

	fmt.Println(result.Message)
	return nil
}