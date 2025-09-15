package cmd

import (
	"fmt"
	"path/filepath"

	"cctools/internal/models"
	"cctools/pkg/fileops"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a file with optional backup",
	Long: `Delete a file with optional backup creation for safety.

The command can create a backup before deletion to allow recovery if needed.
This is especially useful when working with important files that might need
to be restored later.

Examples:
  cctools delete --file unwanted.txt
  cctools delete --file old_config.ini --backup
  cctools delete -f temporary.log --backup --backup-path /safe/backup.log`,
	RunE: runDeleteCmd,
}

var (
	deleteFilePath   string
	deleteBackup     bool
	deleteBackupPath string
)

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringVarP(&deleteFilePath, "file", "f", "", "File path to delete (required)")
	deleteCmd.Flags().BoolVarP(&deleteBackup, "backup", "b", false, "Create backup before deletion")
	deleteCmd.Flags().StringVar(&deleteBackupPath, "backup-path", "", "Custom backup file path")

	deleteCmd.MarkFlagRequired("file")
}

func runDeleteCmd(cmd *cobra.Command, args []string) error {
	// Convert to absolute path
	absFilePath, err := filepath.Abs(deleteFilePath)
	if err != nil {
		return fmt.Errorf("failed to resolve file path: %w", err)
	}

	// Create file operations handler
	fileOps := fileops.NewFileOperations()

	// Get verbose flag from parent command
	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		fmt.Printf("Deleting file: %s\n", absFilePath)
		if deleteBackup {
			if deleteBackupPath != "" {
				fmt.Printf("Backup will be created at: %s\n", deleteBackupPath)
			} else {
				fmt.Println("Backup will be created automatically")
			}
		}
	}

	// Create delete operation
	operation := &models.DeleteOperation{
		FilePath:     absFilePath,
		CreateBackup: deleteBackup,
		BackupPath:   deleteBackupPath,
	}

	// Execute delete operation
	result, err := fileOps.DeleteFile(operation)
	if err != nil {
		return fmt.Errorf("delete failed: %s", result.Message)
	}

	// Display results
	if verbose {
		fmt.Printf("Original encoding: %s\n", result.SourceInfo.Encoding)
		fmt.Printf("File size: %d bytes\n", len(result.SourceInfo.Content))
		if result.BackupPath != "" {
			fmt.Printf("Backup location: %s\n", result.BackupPath)
		}
	}

	fmt.Println(result.Message)
	return nil
}