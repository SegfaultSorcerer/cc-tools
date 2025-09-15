package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"cctools/internal/models"
	"cctools/pkg/fileops"

	"github.com/spf13/cobra"
)

var multiEditCmd = &cobra.Command{
	Use:   "multiedit",
	Short: "Perform multiple edit operations on a file atomically",
	Long: `Perform multiple edit operations on a single file atomically.

All edits are applied sequentially. If any edit fails, all changes are
rolled back and the original file is restored.

The edits are specified in a JSON file with the following format:
{
  "file_path": "/path/to/file.txt",
  "edits": [
    {
      "old_string": "text to replace",
      "new_string": "replacement text",
      "replace_all": false
    },
    {
      "old_string": "another text",
      "new_string": "another replacement",
      "replace_all": true
    }
  ]
}

Examples:
  cctools multiedit --edits-file edits.json
  cctools multiedit -e my-edits.json --verbose`,
	RunE: runMultiEditCmd,
}

var (
	multiEditFile string
)

func init() {
	rootCmd.AddCommand(multiEditCmd)

	multiEditCmd.Flags().StringVarP(&multiEditFile, "edits-file", "e", "", "JSON file containing edit operations (required)")
	multiEditCmd.MarkFlagRequired("edits-file")
}

// fixWindowsPathsInJSON attempts to fix Windows paths that aren't properly escaped in JSON
func fixWindowsPathsInJSON(data []byte) []byte {
	content := string(data)

	// Regex para encontrar paths Windows mal escapados em file_path
	// Procura por padrões como "C:\Path" e os converte para "C:\\Path"
	pathRegex := regexp.MustCompile(`"file_path"\s*:\s*"([C-Z]:\\[^"]*)"`)

	content = pathRegex.ReplaceAllStringFunc(content, func(match string) string {
		// Extrai o path da string
		parts := strings.Split(match, `"`)
		if len(parts) >= 4 {
			path := parts[3]
			// Escapa as barras invertidas que não estão já escapadas
			fixedPath := strings.ReplaceAll(path, `\`, `\\`)
			return fmt.Sprintf(`"file_path": "%s"`, fixedPath)
		}
		return match
	})

	return []byte(content)
}

func runMultiEditCmd(cmd *cobra.Command, args []string) error {
	// Read the edits file
	editsData, err := os.ReadFile(multiEditFile)
	if err != nil {
		return fmt.Errorf("failed to read edits file '%s': %w", multiEditFile, err)
	}

	// Parse the JSON
	var editRequest models.MultiEditRequest
	if err := json.Unmarshal(editsData, &editRequest); err != nil {
		// Se falhou o parsing, tenta corrigir paths Windows mal escapados
		if strings.Contains(err.Error(), "invalid character") && strings.Contains(err.Error(), "in string escape code") {
			fixedData := fixWindowsPathsInJSON(editsData)
			if err2 := json.Unmarshal(fixedData, &editRequest); err2 != nil {
				return fmt.Errorf("failed to parse edits file even after attempting Windows path fix.\nOriginal error: %w\nSuggestion: ensure Windows paths use double backslashes (e.g., \"C:\\\\path\\\\to\\\\file\")", err)
			}
			// Se conseguiu corrigir, informa ao usuário
			fmt.Println("Info: Windows path automatically corrected in JSON")
		} else {
			return fmt.Errorf("failed to parse edits file: %w", err)
		}
	}

	// Validate the request
	if editRequest.FilePath == "" {
		return fmt.Errorf("file_path is required in edits file")
	}

	if len(editRequest.Edits) == 0 {
		return fmt.Errorf("at least one edit operation is required")
	}

	// Normalize path separators for cross-platform compatibility
	normalizedPath := filepath.FromSlash(editRequest.FilePath)

	// Convert to absolute path if relative
	absPath, err := filepath.Abs(normalizedPath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}
	editRequest.FilePath = absPath

	// Validate each edit
	for i, edit := range editRequest.Edits {
		if edit.OldString == "" {
			return fmt.Errorf("old_string is required for edit %d", i+1)
		}
		if edit.OldString == edit.NewString {
			return fmt.Errorf("old_string and new_string must be different for edit %d", i+1)
		}
	}

	// Create file operations handler
	fileOps := fileops.NewFileOperations()

	// Perform the multi-edit
	result, err := fileOps.MultiEditFile(&editRequest)
	if err != nil {
		return fmt.Errorf("multi-edit failed: %w", err)
	}

	if !result.Success {
		// Display the detailed message
		fmt.Println("Multi-edit operation failed:")
		fmt.Println(result.Message)
		return fmt.Errorf("operation unsuccessful")
	}

	// Get verbose flag from parent command
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		fmt.Printf("Multi-edit completed successfully: %s\n", absPath)
		fmt.Printf("Operation: %s\n", result.Message)
		fmt.Printf("Number of edits: %d\n", len(editRequest.Edits))
		fmt.Println("Edit operations:")
		for i, edit := range editRequest.Edits {
			fmt.Printf("  %d. Replace %q with %q (replace_all: %t)\n",
				i+1, edit.OldString, edit.NewString, edit.ReplaceAll)
		}
	} else {
		fmt.Printf("Multi-edit completed successfully: %s\n", absPath)
		fmt.Printf("%s\n", result.Message)
	}

	return nil
}