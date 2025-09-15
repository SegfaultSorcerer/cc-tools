package cmd

import (
	"fmt"
	"path/filepath"

	"cctools/internal/models"
	"cctools/pkg/fileops"

	"github.com/spf13/cobra"
)

var rmdirCmd = &cobra.Command{
	Use:   "rmdir",
	Short: "Remove directories with optional backup",
	Long: `Remove directories with optional backup creation for safety.
The command can remove empty directories or recursively remove
directory trees with all their contents.

When backup is enabled, the entire directory structure is copied
before deletion, allowing complete recovery if needed.

Examples:
  cctools rmdir --path emptydir
  cctools rmdir --path project/ --recursive
  cctools rmdir --path oldproject/ --recursive --backup
  cctools rmdir -p temp/ --recursive --backup --backup-path /safe/temp_backup/`,
	RunE: runRmdirCmd,
}

var (
	rmdirPath       string
	rmdirRecursive  bool
	rmdirBackup     bool
	rmdirBackupPath string
)

func init() {
	rootCmd.AddCommand(rmdirCmd)

	rmdirCmd.Flags().StringVarP(&rmdirPath, "path", "p", "", "Directory path to remove (required)")
	rmdirCmd.Flags().BoolVarP(&rmdirRecursive, "recursive", "r", false, "Remove directories and their contents recursively")
	rmdirCmd.Flags().BoolVarP(&rmdirBackup, "backup", "b", false, "Create backup before deletion")
	rmdirCmd.Flags().StringVar(&rmdirBackupPath, "backup-path", "", "Custom backup directory path")

	rmdirCmd.MarkFlagRequired("path")
}

func runRmdirCmd(cmd *cobra.Command, args []string) error {
	// Convert to absolute path
	absPath, err := filepath.Abs(rmdirPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Create file operations handler
	fileOps := fileops.NewFileOperations()

	// Get verbose flag from parent command
	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		fmt.Printf("Removing directory: %s\n", absPath)
		if rmdirRecursive {
			fmt.Println("Will remove directory and all contents recursively")
		} else {
			fmt.Println("Will only remove if directory is empty")
		}
		if rmdirBackup {
			if rmdirBackupPath != "" {
				fmt.Printf("Backup will be created at: %s\n", rmdirBackupPath)
			} else {
				fmt.Println("Backup will be created automatically")
			}
		}
	}

	// Create delete operation
	operation := &models.DirectoryDeleteOperation{
		Path:         absPath,
		Recursive:    rmdirRecursive,
		CreateBackup: rmdirBackup,
		BackupPath:   rmdirBackupPath,
	}

	// Execute delete operation
	result, err := fileOps.DeleteDirectory(operation)
	if err != nil {
		return fmt.Errorf("rmdir failed: %s", result.Message)
	}

	// Display results
	if verbose {
		if result.ProcessedFiles > 0 {
			fmt.Printf("Files backed up: %d\n", result.ProcessedFiles)
			fmt.Printf("Directories backed up: %d\n", result.ProcessedDirs)
		}
		if result.BackupPath != "" {
			fmt.Printf("Backup location: %s\n", result.BackupPath)
		}
	}

	fmt.Println(result.Message)
	return nil
}