package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"cctools/internal/models"
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
	editFilePath        string
	editOldString       string
	editNewString       string
	editReplaceAll      bool
	editPreview         bool
	editUseRegex        bool
	editFuzzyMatch      bool
	editIgnoreWhitespace bool
	editCaseInsensitive bool
)

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.Flags().StringVarP(&editFilePath, "file", "f", "", "Path to the file to edit (required)")
	editCmd.Flags().StringVarP(&editOldString, "old", "o", "", "String to be replaced (required)")
	editCmd.Flags().StringVarP(&editNewString, "new", "n", "", "Replacement string (required)")
	editCmd.Flags().BoolVar(&editReplaceAll, "replace-all", false, "Replace all occurrences (default: false)")
	editCmd.Flags().BoolVar(&editPreview, "preview", false, "Preview changes without applying them")
	editCmd.Flags().BoolVar(&editUseRegex, "regex", false, "Treat old string as regular expression")
	editCmd.Flags().BoolVar(&editFuzzyMatch, "fuzzy", false, "Enable fuzzy matching for strings")
	editCmd.Flags().BoolVar(&editIgnoreWhitespace, "ignore-whitespace", false, "Ignore differences in whitespace")
	editCmd.Flags().BoolVar(&editCaseInsensitive, "case-insensitive", false, "Perform case-insensitive matching")

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

	// Setup matching options
	options := &models.MatchingOptions{
		UseRegex:         editUseRegex,
		FuzzyMatch:       editFuzzyMatch,
		IgnoreWhitespace: editIgnoreWhitespace,
		CaseInsensitive:  editCaseInsensitive,
	}

	// Perform the edit with advanced options
	result, err := fileOps.EditFileWithOptions(absPath, editOldString, editNewString, editReplaceAll, options, editPreview)
	if err != nil {
		return fmt.Errorf("edit failed: %w", err)
	}

	if !result.Success {
		// Display the detailed message
		fmt.Println("Edit operation failed:")
		fmt.Println(result.Message)

		// Show matching suggestions if available
		if len(result.MatchedLines) > 0 {
			fmt.Println("\nSimilar matches found:")
			for _, match := range result.MatchedLines {
				fmt.Printf("Line %d: %s\n", match.LineNumber, strings.TrimSpace(match.LineText))
			}
		}

		return fmt.Errorf("operation unsuccessful")
	}

	// Handle preview mode
	if editPreview {
		fmt.Println("PREVIEW MODE - No changes applied")
		fmt.Println(result.PreviewDiff)

		if len(result.MatchedLines) > 0 {
			fmt.Println("Matches found:")
			for _, match := range result.MatchedLines {
				fmt.Printf("  Line %d: %s\n", match.LineNumber, strings.TrimSpace(match.LineText))
				if len(match.Context) > 0 {
					fmt.Println("  Context:")
					for _, contextLine := range match.Context {
						fmt.Printf("    %s\n", contextLine)
					}
				}
			}
		}
		return nil
	}

	// Get verbose flag from parent command
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		fmt.Printf("Edit completed successfully: %s\n", absPath)
		fmt.Printf("Operation: %s\n", result.Message)
		fmt.Printf("Replace all: %t\n", editReplaceAll)
		fmt.Printf("Old string: %q\n", editOldString)
		fmt.Printf("New string: %q\n", editNewString)
		if editUseRegex {
			fmt.Printf("Regex mode: enabled\n")
		}
		if editFuzzyMatch {
			fmt.Printf("Fuzzy matching: enabled\n")
		}
		if len(result.MatchedLines) > 0 {
			fmt.Printf("Matches processed: %d\n", len(result.MatchedLines))
		}
	} else {
		fmt.Printf("File edited successfully: %s\n", absPath)
	}

	return nil
}