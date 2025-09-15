package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"cctools/pkg/fileops"

	"github.com/spf13/cobra"
)

var previewCmd = &cobra.Command{
	Use:   "preview",
	Short: "Preview edit operations without modifying the file",
	Long: `Preview what would be changed by an edit operation without actually modifying the file.

This is useful for debugging and verifying that your edit strings are correct
before performing the actual operation.

The command shows:
- The original file encoding
- Exact matches and their context
- Similar matches if no exact match is found
- Preview of the changes that would be made

Examples:
  cctools preview --file /path/to/file.txt --old "hello" --new "hi"
  cctools preview --file file.txt --old "debug: false" --new "debug: true"
  cctools preview -f file.txt -o "old text" -n "new text" --replace-all`,
	RunE: runPreviewCmd,
}

var (
	previewFilePath   string
	previewOldString  string
	previewNewString  string
	previewReplaceAll bool
)

func init() {
	rootCmd.AddCommand(previewCmd)

	previewCmd.Flags().StringVarP(&previewFilePath, "file", "f", "", "Path to the file to preview (required)")
	previewCmd.Flags().StringVarP(&previewOldString, "old", "o", "", "String to be replaced (required)")
	previewCmd.Flags().StringVarP(&previewNewString, "new", "n", "", "Replacement string (required)")
	previewCmd.Flags().BoolVar(&previewReplaceAll, "replace-all", false, "Preview replacing all occurrences (default: false)")

	previewCmd.MarkFlagRequired("file")
	previewCmd.MarkFlagRequired("old")
	previewCmd.MarkFlagRequired("new")
}

func runPreviewCmd(cmd *cobra.Command, args []string) error {
	// Convert to absolute path if relative
	absPath, err := filepath.Abs(previewFilePath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Create file operations handler
	fileOps := fileops.NewFileOperations()

	// Read file with encoding detection
	fileInfo, err := fileOps.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	fmt.Printf("Preview for file: %s\n", absPath)
	fmt.Printf("Detected encoding: %s\n", fileInfo.Encoding)
	fmt.Printf("Old string: %q\n", previewOldString)
	fmt.Printf("New string: %q\n", previewNewString)
	fmt.Printf("Replace all: %t\n", previewReplaceAll)
	fmt.Println()

	// Decode content for analysis
	content, err := fileOps.DecodeFileContent(fileInfo)
	if err != nil {
		return fmt.Errorf("failed to decode file content: %w", err)
	}

	// Find matches
	count := strings.Count(content, previewOldString)
	fmt.Printf("Exact matches found: %d\n", count)

	if count == 0 {
		fmt.Println("\n❌ No exact matches found")

		// Show similar matches
		matches := fileOps.FindSimilarMatches(content, previewOldString)
		if len(matches) > 0 {
			fmt.Println("\n🔍 Similar matches found:")
			for _, match := range matches {
				fmt.Printf("  %s\n", match)
			}
		}

		// Try fuzzy matching
		if lineIndex, matchedLine := fileOps.FindBestMatch(content, previewOldString); lineIndex != -1 {
			fmt.Printf("\n💡 Best fuzzy match at line %d: %q\n", lineIndex+1, matchedLine)
		}

		return nil
	}

	// Show match context
	lines := strings.Split(content, "\n")
	fmt.Println("\n📍 Match locations:")
	for i, line := range lines {
		if strings.Contains(line, previewOldString) {
			fmt.Printf("  Line %d: %q\n", i+1, strings.TrimSpace(line))
		}
	}

	// Show preview of changes
	var newContent string
	if previewReplaceAll || count == 1 {
		if previewReplaceAll {
			newContent = strings.ReplaceAll(content, previewOldString, previewNewString)
			fmt.Printf("\n✅ Would replace %d occurrences\n", count)
		} else {
			newContent = strings.Replace(content, previewOldString, previewNewString, 1)
			fmt.Println("\n✅ Would replace 1 occurrence")
		}

		// Show before/after context
		fmt.Println("\n📋 Preview of changes:")
		for i, line := range lines {
			newLines := strings.Split(newContent, "\n")
			if i < len(newLines) && line != newLines[i] {
				fmt.Printf("  Line %d:\n", i+1)
				fmt.Printf("    Before: %q\n", strings.TrimSpace(line))
				fmt.Printf("    After:  %q\n", strings.TrimSpace(newLines[i]))
			}
		}
	} else {
		fmt.Printf("\n⚠️  Multiple matches found (%d), use --replace-all to replace all\n", count)
		fmt.Println("    Only showing first match for preview:")

		firstMatch := strings.Index(content, previewOldString)
		lineNumber := strings.Count(content[:firstMatch], "\n") + 1
		fmt.Printf("    Line %d would be changed\n", lineNumber)
	}

	return nil
}