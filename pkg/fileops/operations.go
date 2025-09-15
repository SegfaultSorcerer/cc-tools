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

// CopyFile copies a file from source to destination, preserving encoding
func (f *FileOperations) CopyFile(operation *models.CopyOperation) (*models.FileOperationResult, error) {
	// Validate source file exists
	if _, err := os.Stat(operation.SourcePath); os.IsNotExist(err) {
		return &models.FileOperationResult{
			Success: false,
			Message: "Source file does not exist",
			Error:   err,
		}, err
	}

	// Check if destination already exists and overwrite is not enabled
	if _, err := os.Stat(operation.DestinationPath); err == nil && !operation.Overwrite {
		return &models.FileOperationResult{
			Success: false,
			Message: "Destination file already exists, use overwrite flag to replace",
			Error:   fmt.Errorf("destination exists"),
		}, fmt.Errorf("destination exists")
	}

	// Read source file with encoding detection
	sourceInfo, err := f.ReadFile(operation.SourcePath)
	if err != nil {
		return &models.FileOperationResult{
			Success: false,
			Message: "Failed to read source file",
			Error:   err,
		}, err
	}

	// Ensure destination directory exists
	destDir := filepath.Dir(operation.DestinationPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return &models.FileOperationResult{
			Success: false,
			Message: "Failed to create destination directory",
			Error:   err,
		}, err
	}

	// Copy file preserving original encoding
	if err := os.WriteFile(operation.DestinationPath, sourceInfo.Content, 0644); err != nil {
		return &models.FileOperationResult{
			Success: false,
			Message: "Failed to write destination file",
			Error:   err,
		}, err
	}

	// If preserve mode is enabled, copy file permissions
	if operation.PreserveMode {
		sourceStats, err := os.Stat(operation.SourcePath)
		if err == nil {
			os.Chmod(operation.DestinationPath, sourceStats.Mode())
		}
	}

	// Read destination file info for result
	destInfo, _ := f.ReadFile(operation.DestinationPath)

	return &models.FileOperationResult{
		Success:    true,
		Message:    fmt.Sprintf("File copied successfully from %s to %s", operation.SourcePath, operation.DestinationPath),
		SourceInfo: sourceInfo,
		TargetInfo: destInfo,
	}, nil
}

// MoveFile moves a file from source to destination, preserving encoding
func (f *FileOperations) MoveFile(operation *models.MoveOperation) (*models.FileOperationResult, error) {
	// Validate source file exists
	if _, err := os.Stat(operation.SourcePath); os.IsNotExist(err) {
		return &models.FileOperationResult{
			Success: false,
			Message: "Source file does not exist",
			Error:   err,
		}, err
	}

	// Check if destination already exists and overwrite is not enabled
	if _, err := os.Stat(operation.DestinationPath); err == nil && !operation.Overwrite {
		return &models.FileOperationResult{
			Success: false,
			Message: "Destination file already exists, use overwrite flag to replace",
			Error:   fmt.Errorf("destination exists"),
		}, fmt.Errorf("destination exists")
	}

	// Read source file info before moving
	sourceInfo, err := f.ReadFile(operation.SourcePath)
	if err != nil {
		return &models.FileOperationResult{
			Success: false,
			Message: "Failed to read source file",
			Error:   err,
		}, err
	}

	// Create backup of source for rollback
	backupPath := operation.SourcePath + ".move_backup"
	if err := f.createBackup(operation.SourcePath, backupPath); err != nil {
		return &models.FileOperationResult{
			Success: false,
			Message: "Failed to create backup before move",
			Error:   err,
		}, err
	}

	// Ensure destination directory exists
	destDir := filepath.Dir(operation.DestinationPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		os.Remove(backupPath)
		return &models.FileOperationResult{
			Success: false,
			Message: "Failed to create destination directory",
			Error:   err,
		}, err
	}

	// Try direct rename first (most efficient if on same filesystem)
	if err := os.Rename(operation.SourcePath, operation.DestinationPath); err != nil {
		// If rename fails, do copy + delete
		copyOp := &models.CopyOperation{
			SourcePath:      operation.SourcePath,
			DestinationPath: operation.DestinationPath,
			PreserveMode:    true,
			Overwrite:       operation.Overwrite,
		}

		result, err := f.CopyFile(copyOp)
		if err != nil {
			f.restoreBackup(backupPath, operation.SourcePath)
			return result, err
		}

		// Delete source file after successful copy
		if err := os.Remove(operation.SourcePath); err != nil {
			// Copy succeeded but delete failed - clean up destination and restore source
			os.Remove(operation.DestinationPath)
			f.restoreBackup(backupPath, operation.SourcePath)
			return &models.FileOperationResult{
				Success: false,
				Message: "Failed to remove source file after copy",
				Error:   err,
			}, err
		}
	}

	// Remove backup on success
	os.Remove(backupPath)

	// Read destination file info for result
	destInfo, _ := f.ReadFile(operation.DestinationPath)

	return &models.FileOperationResult{
		Success:    true,
		Message:    fmt.Sprintf("File moved successfully from %s to %s", operation.SourcePath, operation.DestinationPath),
		SourceInfo: sourceInfo,
		TargetInfo: destInfo,
	}, nil
}

// DeleteFile deletes a file with optional backup
func (f *FileOperations) DeleteFile(operation *models.DeleteOperation) (*models.FileOperationResult, error) {
	// Validate file exists
	if _, err := os.Stat(operation.FilePath); os.IsNotExist(err) {
		return &models.FileOperationResult{
			Success: false,
			Message: "File does not exist",
			Error:   err,
		}, err
	}

	// Read file info before deletion
	fileInfo, err := f.ReadFile(operation.FilePath)
	if err != nil {
		return &models.FileOperationResult{
			Success: false,
			Message: "Failed to read file before deletion",
			Error:   err,
		}, err
	}

	var backupPath string

	// Create backup if requested
	if operation.CreateBackup {
		if operation.BackupPath != "" {
			backupPath = operation.BackupPath
		} else {
			backupPath = operation.FilePath + ".deleted_backup"
		}

		if err := f.createBackup(operation.FilePath, backupPath); err != nil {
			return &models.FileOperationResult{
				Success: false,
				Message: "Failed to create backup before deletion",
				Error:   err,
			}, err
		}
	}

	// Delete the file
	if err := os.Remove(operation.FilePath); err != nil {
		// If backup was created and deletion failed, remove the backup
		if operation.CreateBackup && backupPath != "" {
			os.Remove(backupPath)
		}
		return &models.FileOperationResult{
			Success: false,
			Message: "Failed to delete file",
			Error:   err,
		}, err
	}

	message := fmt.Sprintf("File deleted successfully: %s", operation.FilePath)
	if operation.CreateBackup && backupPath != "" {
		message += fmt.Sprintf(" (backup created at: %s)", backupPath)
	}

	return &models.FileOperationResult{
		Success:    true,
		Message:    message,
		BackupPath: backupPath,
		SourceInfo: fileInfo,
	}, nil
}

// CreateDirectory creates a directory with optional parent creation
func (f *FileOperations) CreateDirectory(operation *models.DirectoryOperation) (*models.DirectoryOperationResult, error) {
	// Check if directory already exists
	if _, err := os.Stat(operation.Path); err == nil {
		return &models.DirectoryOperationResult{
			Success: false,
			Message: fmt.Sprintf("Directory already exists: %s", operation.Path),
			Error:   fmt.Errorf("directory exists"),
		}, fmt.Errorf("directory exists")
	}

	// Create directory
	mode := os.FileMode(0755)
	if operation.Mode != 0 {
		mode = os.FileMode(operation.Mode)
	}

	var err error
	if operation.CreateParents {
		err = os.MkdirAll(operation.Path, mode)
	} else {
		err = os.Mkdir(operation.Path, mode)
	}

	if err != nil {
		return &models.DirectoryOperationResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create directory: %s", err.Error()),
			Error:   err,
		}, err
	}

	return &models.DirectoryOperationResult{
		Success:       true,
		Message:       fmt.Sprintf("Directory created successfully: %s", operation.Path),
		ProcessedDirs: 1,
	}, nil
}

// CopyDirectory copies a directory recursively preserving encoding
func (f *FileOperations) CopyDirectory(operation *models.DirectoryCopyOperation) (*models.DirectoryOperationResult, error) {
	// Validate source directory exists
	sourceInfo, err := os.Stat(operation.SourcePath)
	if os.IsNotExist(err) {
		return &models.DirectoryOperationResult{
			Success: false,
			Message: "Source directory does not exist",
			Error:   err,
		}, err
	}

	if !sourceInfo.IsDir() {
		return &models.DirectoryOperationResult{
			Success: false,
			Message: "Source path is not a directory",
			Error:   fmt.Errorf("not a directory"),
		}, fmt.Errorf("not a directory")
	}

	// Check if destination exists and handle overwrite
	if _, err := os.Stat(operation.DestinationPath); err == nil && !operation.Overwrite {
		return &models.DirectoryOperationResult{
			Success: false,
			Message: "Destination directory already exists, use overwrite flag",
			Error:   fmt.Errorf("destination exists"),
		}, fmt.Errorf("destination exists")
	}

	// Create destination directory
	if err := os.MkdirAll(operation.DestinationPath, sourceInfo.Mode()); err != nil {
		return &models.DirectoryOperationResult{
			Success: false,
			Message: "Failed to create destination directory",
			Error:   err,
		}, err
	}

	var processedFiles, processedDirs int
	var totalSize int64
	sourceEncoding := make(map[string]int)
	targetEncoding := make(map[string]int)

	// Walk through source directory
	err = filepath.Walk(operation.SourcePath, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(operation.SourcePath, srcPath)
		if err != nil {
			return err
		}

		destPath := filepath.Join(operation.DestinationPath, relPath)

		if info.IsDir() {
			// Create directory
			if err := os.MkdirAll(destPath, info.Mode()); err != nil {
				return err
			}
			processedDirs++
		} else {
			// Check if we should skip existing files
			if operation.SkipExisting {
				if _, err := os.Stat(destPath); err == nil {
					return nil // Skip existing file
				}
			}

			// Read source file with encoding detection
			fileInfo, err := f.ReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", srcPath, err)
			}

			// Track encoding
			sourceEncoding[fileInfo.Encoding]++

			// Ensure destination directory exists
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return err
			}

			// Copy file preserving encoding
			if err := os.WriteFile(destPath, fileInfo.Content, info.Mode()); err != nil {
				return fmt.Errorf("failed to write file %s: %w", destPath, err)
			}

			// Preserve timestamps if requested
			if operation.PreserveAll {
				os.Chtimes(destPath, info.ModTime(), info.ModTime())
			}

			// Track target encoding (should be same as source)
			targetEncoding[fileInfo.Encoding]++
			processedFiles++
			totalSize += int64(len(fileInfo.Content))
		}

		return nil
	})

	if err != nil {
		return &models.DirectoryOperationResult{
			Success: false,
			Message: fmt.Sprintf("Copy failed: %s", err.Error()),
			Error:   err,
		}, err
	}

	return &models.DirectoryOperationResult{
		Success:        true,
		Message:        fmt.Sprintf("Directory copied successfully from %s to %s (%d files, %d directories)", operation.SourcePath, operation.DestinationPath, processedFiles, processedDirs),
		ProcessedFiles: processedFiles,
		ProcessedDirs:  processedDirs,
		TotalSize:      totalSize,
		SourceInfo: &models.DirectoryInfo{
			Path:       operation.SourcePath,
			TotalFiles: processedFiles,
			TotalSize:  totalSize,
			Encodings:  sourceEncoding,
		},
		TargetInfo: &models.DirectoryInfo{
			Path:       operation.DestinationPath,
			TotalFiles: processedFiles,
			TotalSize:  totalSize,
			Encodings:  targetEncoding,
		},
	}, nil
}

// MoveDirectory moves a directory with rollback support
func (f *FileOperations) MoveDirectory(operation *models.DirectoryMoveOperation) (*models.DirectoryOperationResult, error) {
	// Try direct rename first (most efficient)
	if err := os.Rename(operation.SourcePath, operation.DestinationPath); err == nil {
		return &models.DirectoryOperationResult{
			Success: true,
			Message: fmt.Sprintf("Directory moved successfully from %s to %s", operation.SourcePath, operation.DestinationPath),
		}, nil
	}

	// If rename fails, do copy + delete with rollback
	copyOp := &models.DirectoryCopyOperation{
		SourcePath:      operation.SourcePath,
		DestinationPath: operation.DestinationPath,
		PreserveAll:     true,
		Overwrite:       operation.Overwrite,
	}

	result, err := f.CopyDirectory(copyOp)
	if err != nil {
		return result, err
	}

	// Remove source directory after successful copy
	if err := os.RemoveAll(operation.SourcePath); err != nil {
		// Copy succeeded but delete failed - clean up destination
		os.RemoveAll(operation.DestinationPath)
		return &models.DirectoryOperationResult{
			Success: false,
			Message: "Failed to remove source directory after copy",
			Error:   err,
		}, err
	}

	result.Message = fmt.Sprintf("Directory moved successfully from %s to %s (%d files, %d directories)", operation.SourcePath, operation.DestinationPath, result.ProcessedFiles, result.ProcessedDirs)
	return result, nil
}

// DeleteDirectory removes a directory with optional backup
func (f *FileOperations) DeleteDirectory(operation *models.DirectoryDeleteOperation) (*models.DirectoryOperationResult, error) {
	// Validate directory exists
	dirInfo, err := os.Stat(operation.Path)
	if os.IsNotExist(err) {
		return &models.DirectoryOperationResult{
			Success: false,
			Message: "Directory does not exist",
			Error:   err,
		}, err
	}

	if !dirInfo.IsDir() {
		return &models.DirectoryOperationResult{
			Success: false,
			Message: "Path is not a directory",
			Error:   fmt.Errorf("not a directory"),
		}, fmt.Errorf("not a directory")
	}

	var backupPath string
	var processedFiles, processedDirs int

	// Create backup if requested
	if operation.CreateBackup {
		if operation.BackupPath != "" {
			backupPath = operation.BackupPath
		} else {
			backupPath = operation.Path + ".deleted_backup"
		}

		copyOp := &models.DirectoryCopyOperation{
			SourcePath:      operation.Path,
			DestinationPath: backupPath,
			PreserveAll:     true,
			Overwrite:       false,
		}

		result, err := f.CopyDirectory(copyOp)
		if err != nil {
			return &models.DirectoryOperationResult{
				Success: false,
				Message: "Failed to create backup before deletion",
				Error:   err,
			}, err
		}
		processedFiles = result.ProcessedFiles
		processedDirs = result.ProcessedDirs
	}

	// Delete the directory
	if operation.Recursive {
		err = os.RemoveAll(operation.Path)
	} else {
		err = os.Remove(operation.Path)
	}

	if err != nil {
		// If backup was created and deletion failed, remove the backup
		if operation.CreateBackup && backupPath != "" {
			os.RemoveAll(backupPath)
		}
		return &models.DirectoryOperationResult{
			Success: false,
			Message: "Failed to delete directory",
			Error:   err,
		}, err
	}

	message := fmt.Sprintf("Directory deleted successfully: %s", operation.Path)
	if operation.CreateBackup && backupPath != "" {
		message += fmt.Sprintf(" (backup created at: %s)", backupPath)
	}

	return &models.DirectoryOperationResult{
		Success:        true,
		Message:        message,
		BackupPath:     backupPath,
		ProcessedFiles: processedFiles,
		ProcessedDirs:  processedDirs,
	}, nil
}

// ListDirectory lists directory contents with encoding detection
func (f *FileOperations) ListDirectory(operation *models.DirectoryListOperation) (*models.DirectoryOperationResult, error) {
	// Validate directory exists
	if _, err := os.Stat(operation.Path); os.IsNotExist(err) {
		return &models.DirectoryOperationResult{
			Success: false,
			Message: "Directory does not exist",
			Error:   err,
		}, err
	}

	var fileList []models.FileEntry
	var processedFiles, processedDirs int
	encodings := make(map[string]int)
	fileTypes := make(map[string]int)

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden files unless requested
		if !operation.ShowHidden && strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Apply filter if specified
		if operation.Filter != "" {
			matched, err := filepath.Match(operation.Filter, info.Name())
			if err != nil || !matched {
				if info.IsDir() && !operation.Recursive {
					return filepath.SkipDir
				}
				return nil
			}
		}

		entry := models.FileEntry{
			Path:  path,
			Name:  info.Name(),
			Size:  info.Size(),
			IsDir: info.IsDir(),
			Mode:  info.Mode().String(),
		}

		if info.IsDir() {
			processedDirs++
		} else {
			processedFiles++

			// Get file extension
			ext := filepath.Ext(info.Name())
			fileTypes[ext]++

			// Detect encoding if requested
			if operation.ShowEncoding {
				if fileInfo, err := f.ReadFile(path); err == nil {
					entry.Encoding = fileInfo.Encoding
					encodings[fileInfo.Encoding]++
				}
			}
		}

		fileList = append(fileList, entry)

		// If not recursive and this is a directory, skip its contents
		if !operation.Recursive && info.IsDir() && path != operation.Path {
			return filepath.SkipDir
		}

		return nil
	}

	var err error
	if operation.Recursive {
		err = filepath.Walk(operation.Path, walkFunc)
	} else {
		entries, err := os.ReadDir(operation.Path)
		if err != nil {
			return &models.DirectoryOperationResult{
				Success: false,
				Message: "Failed to read directory",
				Error:   err,
			}, err
		}

		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			walkFunc(filepath.Join(operation.Path, entry.Name()), info, nil)
		}
	}

	if err != nil {
		return &models.DirectoryOperationResult{
			Success: false,
			Message: fmt.Sprintf("Failed to list directory: %s", err.Error()),
			Error:   err,
		}, err
	}

	return &models.DirectoryOperationResult{
		Success:        true,
		Message:        fmt.Sprintf("Listed %d files and %d directories", processedFiles, processedDirs),
		ProcessedFiles: processedFiles,
		ProcessedDirs:  processedDirs,
		FileList:       fileList,
		SourceInfo: &models.DirectoryInfo{
			Path:       operation.Path,
			TotalFiles: processedFiles,
			Encodings:  encodings,
			FileTypes:  fileTypes,
		},
	}, nil
}