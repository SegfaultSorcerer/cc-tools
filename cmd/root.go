package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cctools",
	Short: "CLI file editing tools with encoding support",
	Long: `CCTools is a command-line interface for file editing operations
that preserves the original encoding of files during edits.

It provides commands for:
- Reading files with encoding detection
- Writing files with specified encoding
- Editing files with string replacement
- Multiple atomic edits on a single file
- Previewing edits before applying them
- Copying files while preserving encoding
- Moving files while preserving encoding
- Deleting files with optional backup
- Creating directories with proper permissions
- Copying directories recursively with encoding preservation
- Moving directories with atomic rollback support
- Removing directories with optional backup
- Listing directory contents with encoding analysis

All file and directory operations preserve the original file encoding automatically.`,
	Version: "1.0.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Add global flags here if needed
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
}