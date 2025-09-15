package cmd

import (
	"fmt"
	"path/filepath"
	"strconv"

	"cctools/internal/models"
	"cctools/pkg/fileops"

	"github.com/spf13/cobra"
)

var mkdirCmd = &cobra.Command{
	Use:   "mkdir",
	Short: "Create directories with optional parent creation",
	Long: `Create directories with support for creating parent directories automatically.

The command can create single directories or complete directory structures
with customizable permissions.

Examples:
  cctools mkdir --path newdir
  cctools mkdir --path deep/nested/structure --parents
  cctools mkdir -p project/src/main --parents --mode 755`,
	RunE: runMkdirCmd,
}

var (
	mkdirPath    string
	mkdirParents bool
	mkdirMode    string
)

func init() {
	rootCmd.AddCommand(mkdirCmd)

	mkdirCmd.Flags().StringVarP(&mkdirPath, "path", "p", "", "Directory path to create (required)")
	mkdirCmd.Flags().BoolVar(&mkdirParents, "parents", false, "Create parent directories as needed")
	mkdirCmd.Flags().StringVar(&mkdirMode, "mode", "755", "Directory permissions (octal)")

	mkdirCmd.MarkFlagRequired("path")
}

func runMkdirCmd(cmd *cobra.Command, args []string) error {
	// Convert to absolute path
	absPath, err := filepath.Abs(mkdirPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Parse mode
	mode := 0755
	if mkdirMode != "" {
		parsedMode, err := strconv.ParseInt(mkdirMode, 8, 32)
		if err != nil {
			return fmt.Errorf("invalid mode format: %w", err)
		}
		mode = int(parsedMode)
	}

	// Create file operations handler
	fileOps := fileops.NewFileOperations()

	// Get verbose flag from parent command
	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		fmt.Printf("Creating directory: %s\n", absPath)
		if mkdirParents {
			fmt.Println("Will create parent directories as needed")
		}
		fmt.Printf("Permissions: %o\n", mode)
	}

	// Create directory operation
	operation := &models.DirectoryOperation{
		Path:          absPath,
		CreateParents: mkdirParents,
		Mode:          mode,
	}

	// Execute directory creation
	result, err := fileOps.CreateDirectory(operation)
	if err != nil {
		return fmt.Errorf("mkdir failed: %s", result.Message)
	}

	// Display results
	if verbose {
		fmt.Printf("Directories created: %d\n", result.ProcessedDirs)
	}

	fmt.Println(result.Message)
	return nil
}