package fileops

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"cctools/internal/models"
	"cctools/pkg/encoding"
)

// FileOperations handles file read/write operations with encoding support
type FileOperations struct {
	detector *encoding.Detector
}

// NewFileOperations creates a new file operations handler
func NewFileOperations() *FileOperations {
	return &FileOperations{
		detector: encoding.NewDetector(),
	}
}

// ReadFile reads a file and detects its encoding
func (f *FileOperations) ReadFile(filePath string) (*models.FileInfo, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", filePath)
	}

	// Detect encoding
	detectedEncoding, err := f.detector.DetectFileEncoding(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect encoding: %w", err)
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return &models.FileInfo{
		Path:     filePath,
		Encoding: detectedEncoding,
		Content:  content,
	}, nil
}

// WriteFile writes content to a file with specified encoding
func (f *FileOperations) WriteFile(filePath, content, targetEncoding string) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Encode content
	encodedContent, err := encoding.EncodeString(content, targetEncoding)
	if err != nil {
		return fmt.Errorf("failed to encode content: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, encodedContent, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// EditFile performs a single edit operation on a file
func (f *FileOperations) EditFile(filePath, oldString, newString string, replaceAll bool) (*models.EditResult, error) {
	// Read file with encoding detection
	fileInfo, err := f.ReadFile(filePath)
	if err != nil {
		return &models.EditResult{
			Success: false,
			Message: "Failed to read file",
			Error:   err,
		}, err
	}

	// Decode content to UTF-8 for manipulation
	content, err := encoding.DecodeBytes(fileInfo.Content, fileInfo.Encoding)
	if err != nil {
		return &models.EditResult{
			Success: false,
			Message: "Failed to decode file content",
			Error:   err,
		}, err
	}

	// Perform replacement
	var newContent string
	if replaceAll {
		count := strings.Count(content, oldString)
		if count == 0 {
			// Try to find similar matches for better error message
			matches := findStringMatches(content, oldString)
			errorMsg := fmt.Sprintf("String '%s' not found in file", oldString)
			if len(matches) > 0 {
				errorMsg += "\nSimilar matches found:\n" + strings.Join(matches, "\n")
			}
			return &models.EditResult{
				Success: false,
				Message: errorMsg,
				Error:   fmt.Errorf("old string not found"),
			}, nil
		}
		newContent = strings.ReplaceAll(content, oldString, newString)
	} else {
		// Check if old string exists and is unique
		count := strings.Count(content, oldString)
		if count == 0 {
			// Try to find similar matches for better error message
			matches := findStringMatches(content, oldString)
			errorMsg := fmt.Sprintf("String '%s' not found in file", oldString)
			if len(matches) > 0 {
				errorMsg += "\nSimilar matches found:\n" + strings.Join(matches, "\n")
			}

			// Try fuzzy matching
			if lineIndex, matchedLine := findBestMatch(content, oldString); lineIndex != -1 {
				errorMsg += fmt.Sprintf("\nBest fuzzy match found at line %d: %q", lineIndex+1, matchedLine)
			}

			return &models.EditResult{
				Success: false,
				Message: errorMsg,
				Error:   fmt.Errorf("old string not found"),
			}, nil
		}
		if count > 1 {
			// Show context for all matches
			lines := strings.Split(content, "\n")
			var matchLines []string
			for i, line := range lines {
				if strings.Contains(line, oldString) {
					matchLines = append(matchLines, fmt.Sprintf("Line %d: %q", i+1, strings.TrimSpace(line)))
				}
			}
			errorMsg := fmt.Sprintf("String '%s' appears %d times in file, use --replace-all flag", oldString, count)
			if len(matchLines) > 0 {
				errorMsg += "\nMatches found at:\n" + strings.Join(matchLines, "\n")
			}

			return &models.EditResult{
				Success: false,
				Message: errorMsg,
				Error:   fmt.Errorf("string not unique"),
			}, nil
		}
		newContent = strings.Replace(content, oldString, newString, 1)
	}

	// Create backup
	backupPath := filePath + ".backup"
	if err := f.createBackup(filePath, backupPath); err != nil {
		return &models.EditResult{
			Success: false,
			Message: "Failed to create backup",
			Error:   err,
		}, err
	}

	// Write the modified content back with original encoding
	if err := f.WriteFile(filePath, newContent, fileInfo.Encoding); err != nil {
		// Restore from backup
		f.restoreBackup(backupPath, filePath)
		return &models.EditResult{
			Success: false,
			Message: "Failed to write modified content",
			Error:   err,
		}, err
	}

	// Remove backup on success
	os.Remove(backupPath)

	return &models.EditResult{
		Success: true,
		Message: "File edited successfully",
		Error:   nil,
	}, nil
}

// MultiEditFile performs multiple edit operations atomically
func (f *FileOperations) MultiEditFile(request *models.MultiEditRequest) (*models.EditResult, error) {
	// Read file with encoding detection
	fileInfo, err := f.ReadFile(request.FilePath)
	if err != nil {
		return &models.EditResult{
			Success: false,
			Message: "Failed to read file",
			Error:   err,
		}, err
	}

	// Decode content to UTF-8 for manipulation
	content, err := encoding.DecodeBytes(fileInfo.Content, fileInfo.Encoding)
	if err != nil {
		return &models.EditResult{
			Success: false,
			Message: "Failed to decode file content",
			Error:   err,
		}, err
	}

	// Create backup
	backupPath := request.FilePath + ".backup"
	if err := f.createBackup(request.FilePath, backupPath); err != nil {
		return &models.EditResult{
			Success: false,
			Message: "Failed to create backup",
			Error:   err,
		}, err
	}

	// Apply all edits sequentially
	workingContent := content
	for i, edit := range request.Edits {
		if edit.ReplaceAll {
			workingContent = strings.ReplaceAll(workingContent, edit.OldString, edit.NewString)
		} else {
			count := strings.Count(workingContent, edit.OldString)
			if count == 0 {
				f.restoreBackup(backupPath, request.FilePath)

				// Enhanced error message with similar matches
				matches := findStringMatches(workingContent, edit.OldString)
				errorMsg := fmt.Sprintf("Edit %d failed: string '%s' not found", i+1, edit.OldString)
				if len(matches) > 0 {
					errorMsg += "\nSimilar matches found:\n" + strings.Join(matches, "\n")
				}

				// Try fuzzy matching
				if lineIndex, matchedLine := findBestMatch(workingContent, edit.OldString); lineIndex != -1 {
					errorMsg += fmt.Sprintf("\nBest fuzzy match found at line %d: %q", lineIndex+1, matchedLine)
				}

				return &models.EditResult{
					Success: false,
					Message: errorMsg,
					Error:   fmt.Errorf("string not found in edit %d", i+1),
				}, fmt.Errorf("string not found in edit %d", i+1)
			}
			if count > 1 {
				f.restoreBackup(backupPath, request.FilePath)

				// Show context for all matches
				lines := strings.Split(workingContent, "\n")
				var matchLines []string
				for j, line := range lines {
					if strings.Contains(line, edit.OldString) {
						matchLines = append(matchLines, fmt.Sprintf("Line %d: %q", j+1, strings.TrimSpace(line)))
					}
				}
				errorMsg := fmt.Sprintf("Edit %d failed: string '%s' appears %d times, use replace_all: true", i+1, edit.OldString, count)
				if len(matchLines) > 0 {
					errorMsg += "\nMatches found at:\n" + strings.Join(matchLines, "\n")
				}

				return &models.EditResult{
					Success: false,
					Message: errorMsg,
					Error:   fmt.Errorf("string not unique in edit %d", i+1),
				}, fmt.Errorf("string not unique in edit %d", i+1)
			}
			workingContent = strings.Replace(workingContent, edit.OldString, edit.NewString, 1)
		}
	}

	// Write the modified content back with original encoding
	if err := f.WriteFile(request.FilePath, workingContent, fileInfo.Encoding); err != nil {
		// Restore from backup
		f.restoreBackup(backupPath, request.FilePath)
		return &models.EditResult{
			Success: false,
			Message: "Failed to write modified content",
			Error:   err,
		}, err
	}

	// Remove backup on success
	os.Remove(backupPath)

	return &models.EditResult{
		Success: true,
		Message: fmt.Sprintf("Applied %d edits successfully", len(request.Edits)),
		Error:   nil,
	}, nil
}

// createBackup creates a backup of the original file
func (f *FileOperations) createBackup(originalPath, backupPath string) error {
	source, err := os.Open(originalPath)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(backupPath)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

// restoreBackup restores a file from backup
func (f *FileOperations) restoreBackup(backupPath, originalPath string) error {
	return os.Rename(backupPath, originalPath)
}

// normalizeWhitespace normalizes whitespace in a string for better matching
func normalizeWhitespace(s string) string {
	// Replace multiple whitespaces with single spaces
	re := regexp.MustCompile(`\s+`)
	normalized := re.ReplaceAllString(s, " ")
	return strings.TrimSpace(normalized)
}

// findBestMatch attempts to find the best match for a string, considering encoding issues
func findBestMatch(content, target string) (int, string) {
	// First try exact match
	if index := strings.Index(content, target); index != -1 {
		return index, target
	}

	// Try with normalized whitespace
	normalizedTarget := normalizeWhitespace(target)
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		normalizedLine := normalizeWhitespace(line)
		if strings.Contains(normalizedLine, normalizedTarget) {
			return i, line
		}
	}

	// Try fuzzy matching - remove accents and special characters
	simplifiedTarget := simplifyString(target)
	for i, line := range lines {
		simplifiedLine := simplifyString(line)
		if strings.Contains(simplifiedLine, simplifiedTarget) {
			return i, line
		}
	}

	return -1, ""
}

// simplifyString removes accents and special characters for fuzzy matching
func simplifyString(s string) string {
	// Convert to runes for proper unicode handling
	runes := []rune(s)
	result := make([]rune, 0, len(runes))

	for _, r := range runes {
		// Skip combining marks (accents)
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		// Convert to lowercase
		result = append(result, unicode.ToLower(r))
	}

	return string(result)
}

// findStringMatches returns all possible matches with context
func findStringMatches(content, target string) []string {
	var matches []string

	// Exact matches
	count := strings.Count(content, target)
	if count > 0 {
		matches = append(matches, fmt.Sprintf("Exact matches: %d", count))
	}

	// Look for similar strings (lines containing target words)
	targetWords := strings.Fields(target)
	if len(targetWords) > 1 {
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			wordMatches := 0
			for _, word := range targetWords {
				if strings.Contains(line, word) {
					wordMatches++
				}
			}
			if wordMatches > 0 && wordMatches < len(targetWords) {
				matches = append(matches, fmt.Sprintf("Line %d (partial): %q", i+1, strings.TrimSpace(line)))
			}
		}
	}

	return matches
}

// DecodeFileContent is a public wrapper for decoding file content
func (f *FileOperations) DecodeFileContent(fileInfo *models.FileInfo) (string, error) {
	return encoding.DecodeBytes(fileInfo.Content, fileInfo.Encoding)
}

// FindSimilarMatches is a public wrapper for finding similar matches
func (f *FileOperations) FindSimilarMatches(content, target string) []string {
	return findStringMatches(content, target)
}

// FindBestMatch is a public wrapper for finding the best match
func (f *FileOperations) FindBestMatch(content, target string) (int, string) {
	return findBestMatch(content, target)
}