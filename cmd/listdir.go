package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"cctools/internal/models"
	"cctools/pkg/fileops"

	"github.com/spf13/cobra"
)

var listdirCmd = &cobra.Command{
	Use:   "listdir",
	Short: "List directory contents with encoding detection",
	Long: `List directory contents with intelligent encoding detection and filtering.

The command can show file encodings, filter by patterns, and display
directory contents in a tree-like structure. This is especially useful
for analyzing projects with mixed encodings.

Examples:
  cctools listdir --path /project
  cctools listdir --path src/ --recursive --show-encoding
  cctools listdir -p . --filter "*.pas" --show-encoding
  cctools listdir --path /code --recursive --show-hidden`,
	RunE: runListdirCmd,
}

var (
	listdirPath        string
	listdirRecursive   bool
	listdirShowEncoding bool
	listdirFilter      string
	listdirShowHidden  bool
)

func init() {
	rootCmd.AddCommand(listdirCmd)

	listdirCmd.Flags().StringVarP(&listdirPath, "path", "p", ".", "Directory path to list")
	listdirCmd.Flags().BoolVarP(&listdirRecursive, "recursive", "r", false, "List contents recursively")
	listdirCmd.Flags().BoolVar(&listdirShowEncoding, "show-encoding", false, "Show file encoding detection")
	listdirCmd.Flags().StringVar(&listdirFilter, "filter", "", "Filter files by pattern (e.g., *.go)")
	listdirCmd.Flags().BoolVar(&listdirShowHidden, "show-hidden", false, "Show hidden files and directories")
}

func runListdirCmd(cmd *cobra.Command, args []string) error {
	// Convert to absolute path
	absPath, err := filepath.Abs(listdirPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Create file operations handler
	fileOps := fileops.NewFileOperations()

	// Get verbose flag from parent command
	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		fmt.Printf("Listing directory: %s\n", absPath)
		if listdirRecursive {
			fmt.Println("Mode: Recursive")
		}
		if listdirShowEncoding {
			fmt.Println("Will detect and show file encodings")
		}
		if listdirFilter != "" {
			fmt.Printf("Filter: %s\n", listdirFilter)
		}
		if listdirShowHidden {
			fmt.Println("Will show hidden files")
		}
		fmt.Println()
	}

	// Create list operation
	operation := &models.DirectoryListOperation{
		Path:         absPath,
		Recursive:    listdirRecursive,
		ShowEncoding: listdirShowEncoding,
		Filter:       listdirFilter,
		ShowHidden:   listdirShowHidden,
	}

	// Execute list operation
	result, err := fileOps.ListDirectory(operation)
	if err != nil {
		return fmt.Errorf("listdir failed: %s", result.Message)
	}

	// Display results
	fmt.Printf("Directory: %s\n", absPath)

	if len(result.FileList) == 0 {
		fmt.Println("No files found matching criteria.")
		return nil
	}

	// Group entries by directory level for better display
	currentDir := ""
	for _, entry := range result.FileList {
		entryDir := filepath.Dir(entry.Path)

		// Show directory header when we enter a new directory
		if listdirRecursive && entryDir != currentDir {
			currentDir = entryDir
			if entryDir != absPath {
				fmt.Printf("\n%s/:\n", entryDir)
			}
		}

		// Format entry display
		var displayName string
		if listdirRecursive {
			relPath, _ := filepath.Rel(absPath, entry.Path)
			displayName = relPath
		} else {
			displayName = entry.Name
		}

		// Build entry line
		var line strings.Builder

		// Directory indicator
		if entry.IsDir {
			line.WriteString("d")
		} else {
			line.WriteString("-")
		}

		// Permissions
		line.WriteString(fmt.Sprintf(" %-10s", entry.Mode))

		// Size
		if entry.IsDir {
			line.WriteString(" <DIR>        ")
		} else {
			line.WriteString(fmt.Sprintf(" %12d", entry.Size))
		}

		// Encoding (if requested)
		if listdirShowEncoding && !entry.IsDir {
			if entry.Encoding != "" {
				line.WriteString(fmt.Sprintf(" %-12s", entry.Encoding))
			} else {
				line.WriteString(" <unknown>   ")
			}
		}

		// Name
		line.WriteString(" ")
		line.WriteString(displayName)

		fmt.Println(line.String())
	}

	// Summary
	fmt.Printf("\nSummary: %d files, %d directories", result.ProcessedFiles, result.ProcessedDirs)

	if verbose && result.SourceInfo != nil {
		if len(result.SourceInfo.Encodings) > 0 {
			fmt.Println("\nEncodings found:")
			for encoding, count := range result.SourceInfo.Encodings {
				fmt.Printf("  %-15s: %d files\n", encoding, count)
			}
		}

		if len(result.SourceInfo.FileTypes) > 0 {
			fmt.Println("\nFile types:")
			for extension, count := range result.SourceInfo.FileTypes {
				if extension == "" {
					extension = "(no extension)"
				}
				fmt.Printf("  %-15s: %d files\n", extension, count)
			}
		}
	}

	fmt.Println()

	return nil
}