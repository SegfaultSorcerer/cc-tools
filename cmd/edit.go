package cmd

import (
	"fmt"
	"path/filepath"

	"cctools/pkg/fileops"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a file by replacing text strings",
	Long: `Edit a file by replacing old strings with new strings.

The command preserves the original file encoding automatically.
By default, the old string must be unique in the file. Use --replace-all
to replace all occurrences.

The file is backed up before editing and restored if the operation fails.

Examples:
  cctools edit --file /path/to/file.txt --old "hello" --new "hi"
  cctools edit --file file.txt --old "debug: false" --new "debug: true"
  cctools edit -f file.txt -o "old text" -n "new text" --replace-all`,
	RunE: runEditCmd,
}

var (
	editFilePath   string
	editOldString  string
	editNewString  string
	editReplaceAll bool
)

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.Flags().StringVarP(&editFilePath, "file", "f", "", "Path to the file to edit (required)")
	editCmd.Flags().StringVarP(&editOldString, "old", "o", "", "String to be replaced (required)")
	editCmd.Flags().StringVarP(&editNewString, "new", "n", "", "Replacement string (required)")
	editCmd.Flags().BoolVar(&editReplaceAll, "replace-all", false, "Replace all occurrences (default: false)")

	editCmd.MarkFlagRequired("file")
	editCmd.MarkFlagRequired("old")
	editCmd.MarkFlagRequired("new")
}

func runEditCmd(cmd *cobra.Command, args []string) error {
	// Convert to absolute path if relative
	absPath, err := filepath.Abs(editFilePath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Create file operations handler
	fileOps := fileops.NewFileOperations()

	// Perform the edit
	result, err := fileOps.EditFile(absPath, editOldString, editNewString, editReplaceAll)
	if err != nil {
		return fmt.Errorf("edit failed: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("edit operation failed: %s", result.Message)
	}

	// Get verbose flag from parent command
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		fmt.Printf("Edit completed successfully: %s\n", absPath)
		fmt.Printf("Operation: %s\n", result.Message)
		fmt.Printf("Replace all: %t\n", editReplaceAll)
		fmt.Printf("Old string: %q\n", editOldString)
		fmt.Printf("New string: %q\n", editNewString)
	} else {
		fmt.Printf("File edited successfully: %s\n", absPath)
	}

	return nil
}