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
	Short: "Edit a file by replacing text strings with enhanced matching",
	Long: `Edit a file by replacing old strings with new strings using advanced matching strategies.

The command preserves the original file encoding automatically and provides multiple
matching strategies to handle different scenarios:

MATCHING STRATEGIES:
- Exact matching (default): Requires perfect string match
- Fuzzy matching (--fuzzy): Uses similarity-based matching
- Auto-normalize (--auto-normalize): Tolerates whitespace and formatting differences
- Auto-chunk (--auto-chunk): Breaks large strings into smaller chunks for matching
- Regex matching (--regex): Uses regular expressions

ENHANCED FEATURES:
- Automatic whitespace normalization with --auto-normalize
- Configurable similarity threshold with --similarity (0.0-1.0)
- Large string chunking with --auto-chunk and --max-chunk-size
- Detailed preview mode with --preview for safety
- Multi-line and context-aware matching

The file is backed up before editing and restored if the operation fails.

Basic Examples:
  cctools edit --file /path/to/file.txt --old "hello" --new "hi"
  cctools edit --file file.txt --old "debug: false" --new "debug: true"
  cctools edit -f file.txt -o "old text" -n "new text" --replace-all

Advanced Examples:
  cctools edit -f file.txt -o "procedure.*Click" -n "procedure NewClick" --regex
  cctools edit -f file.txt -o "large block" -n "new block" --fuzzy --similarity 0.8
  cctools edit -f file.txt -o "complex procedure" -n "new procedure" --auto-normalize
  cctools edit -f file.txt -o "huge method" -n "new method" --auto-chunk --preview`,
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
	editAutoNormalize   bool
	editSimilarityThreshold float64
	editAutoChunk       bool
	editMaxChunkSize    int
	editSmartCode       bool
	editAggressiveFuzzy bool
	editSmartSuggestions bool
	editCodeLanguage   string
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
	editCmd.Flags().BoolVar(&editAutoNormalize, "auto-normalize", false, "Automatically normalize whitespace and formatting")
	editCmd.Flags().Float64Var(&editSimilarityThreshold, "similarity", 0.7, "Similarity threshold for fuzzy matching (0.0-1.0)")
	editCmd.Flags().BoolVar(&editAutoChunk, "auto-chunk", false, "Automatically break large strings into smaller chunks")
	editCmd.Flags().IntVar(&editMaxChunkSize, "max-chunk-size", 500, "Maximum size for chunks when auto-chunk is enabled")
	editCmd.Flags().BoolVar(&editSmartCode, "smart-code", false, "Enable smart code understanding for better block matching")
	editCmd.Flags().BoolVar(&editAggressiveFuzzy, "aggressive-fuzzy", false, "Enable more aggressive fuzzy matching for irregular formatting")
	editCmd.Flags().BoolVar(&editSmartSuggestions, "smart-suggestions", false, "Enable intelligent suggestions when exact match fails")
	editCmd.Flags().StringVar(&editCodeLanguage, "code-language", "", "Programming language hint (auto-detected from file extension)")

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
		UseRegex:            editUseRegex,
		FuzzyMatch:          editFuzzyMatch,
		IgnoreWhitespace:    editIgnoreWhitespace,
		CaseInsensitive:     editCaseInsensitive,
		AutoNormalize:       editAutoNormalize,
		SimilarityThreshold: editSimilarityThreshold,
		AutoChunk:           editAutoChunk,
		MaxChunkSize:        editMaxChunkSize,
		SmartCode:           editSmartCode,
		AggressiveFuzzy:     editAggressiveFuzzy,
		SmartSuggestions:    editSmartSuggestions,
		CodeLanguage:        editCodeLanguage,
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