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
	return f.EditFileWithOptions(filePath, oldString, newString, replaceAll, &models.MatchingOptions{}, false)
}

// EditFileWithOptions performs a single edit operation on a file with advanced options
func (f *FileOperations) EditFileWithOptions(filePath, oldString, newString string, replaceAll bool, options *models.MatchingOptions, preview bool) (*models.EditResult, error) {
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

	// Try advanced matching first
	matchInfo, matchErr := f.advancedStringMatch(content, oldString, options)

	// If preview mode, show what would be changed
	if preview {
		if matchErr != nil {
			// Try to find similar matches for better preview
			matches := findStringMatches(content, oldString)
			errorMsg := fmt.Sprintf("No matches found for '%s'", oldString)
			if len(matches) > 0 {
				errorMsg += "\nSimilar matches found:\n" + strings.Join(matches, "\n")
			}
			return &models.EditResult{
				Success:     false,
				Message:     errorMsg,
				Error:       matchErr,
				PreviewDiff: "No changes would be made.",
			}, nil
		}

		// Generate preview
		previewDiff := f.generateDiffPreview(matchInfo.MatchText, newString, matchInfo)
		return &models.EditResult{
			Success:      true,
			Message:      "Preview generated successfully",
			PreviewDiff:  previewDiff,
			MatchedLines: []models.MatchInfo{*matchInfo},
		}, nil
	}

	// Perform actual replacement
	var newContent string
	var allMatches []models.MatchInfo

	// Use advanced matching if it found something, otherwise fall back to exact matching
	if matchErr == nil {
		// Use the matched string for replacement
		if replaceAll {
			// Find all matches using advanced matching
			allMatches = f.findAllMatches(content, oldString, options)
			if len(allMatches) == 0 {
				return f.handleNoMatchesFound(content, oldString)
			}
			newContent = f.replaceAllMatches(content, allMatches, newString)
		} else {
			// Single replacement
			allMatches = []models.MatchInfo{*matchInfo}
			newContent = f.replaceSingleMatch(content, matchInfo, newString)
		}
	} else {
		// Fall back to exact string matching
		count := strings.Count(content, oldString)
		if count == 0 {
			return f.handleNoMatchesFound(content, oldString)
		}
		if count > 1 && !replaceAll {
			return f.handleMultipleMatches(content, oldString, count)
		}

		if replaceAll {
			newContent = strings.ReplaceAll(content, oldString, newString)
		} else {
			newContent = strings.Replace(content, oldString, newString, 1)
		}

		// Create match info for exact matches
		allMatches = f.findExactMatches(content, oldString, strings.Split(content, "\n"))
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
		Success:      true,
		Message:      "File edited successfully",
		MatchedLines: allMatches,
		Error:        nil,
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

	// Handle dry run mode
	if request.DryRun {
		return f.performDryRun(content, request)
	}

	// Create backup (only if not dry run)
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
	var allMatches []models.MatchInfo
	var partialErrors []string
	successfulEdits := 0

	for i, edit := range request.Edits {
		// Create matching options for each edit
		options := &models.MatchingOptions{
			UseRegex:         edit.UseRegex,
			FuzzyMatch:       edit.FuzzyMatch,
			IgnoreWhitespace: false, // Could be added to EditOperation if needed
			CaseInsensitive:  false, // Could be added to EditOperation if needed
		}

		// Try advanced matching
		if matchInfo, matchErr := f.advancedStringMatch(workingContent, edit.OldString, options); matchErr == nil {
			// Success - apply the edit
			if edit.ReplaceAll {
				matches := f.findAllMatches(workingContent, edit.OldString, options)
				workingContent = f.replaceAllMatches(workingContent, matches, edit.NewString)
				allMatches = append(allMatches, matches...)
			} else {
				workingContent = f.replaceSingleMatch(workingContent, matchInfo, edit.NewString)
				allMatches = append(allMatches, *matchInfo)
			}
			successfulEdits++
		} else {
			// Edit failed
			errorMsg := fmt.Sprintf("Edit %d failed: string '%s' not found", i+1, edit.OldString)

			// Try to find similar matches for better error message
			matches := findStringMatches(workingContent, edit.OldString)
			if len(matches) > 0 {
				errorMsg += "\nSimilar matches found:\n" + strings.Join(matches, "\n")
			}

			partialErrors = append(partialErrors, errorMsg)

			// If continue on error is disabled, abort everything
			if !request.ContinueOnError {
				f.restoreBackup(backupPath, request.FilePath)
				return &models.EditResult{
					Success:       false,
					Message:       errorMsg,
					Error:         fmt.Errorf("string not found in edit %d", i+1),
					PartialErrors: partialErrors,
				}, fmt.Errorf("string not found in edit %d", i+1)
			}
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

	// Determine success status
	success := len(partialErrors) == 0
	message := fmt.Sprintf("Applied %d/%d edits successfully", successfulEdits, len(request.Edits))

	return &models.EditResult{
		Success:       success,
		Message:       message,
		MatchedLines:  allMatches,
		PartialErrors: partialErrors,
		Error:         nil,
	}, nil
}

// performDryRun simulates the multi-edit operation without making changes
func (f *FileOperations) performDryRun(content string, request *models.MultiEditRequest) (*models.EditResult, error) {
	var allMatches []models.MatchInfo
	var partialErrors []string
	var previewParts []string

	previewParts = append(previewParts, "DRY RUN - Multi-edit Preview")
	previewParts = append(previewParts, "=============================\n")

	for i, edit := range request.Edits {
		previewParts = append(previewParts, fmt.Sprintf("Edit %d: Replace %q with %q", i+1, edit.OldString, edit.NewString))

		// Create matching options for each edit
		options := &models.MatchingOptions{
			UseRegex:         edit.UseRegex,
			FuzzyMatch:       edit.FuzzyMatch,
			IgnoreWhitespace: false,
			CaseInsensitive:  false,
		}

		// Try to find matches
		if matchInfo, matchErr := f.advancedStringMatch(content, edit.OldString, options); matchErr == nil {
			if edit.ReplaceAll {
				matches := f.findAllMatches(content, edit.OldString, options)
				previewParts = append(previewParts, fmt.Sprintf("  ✓ Would replace %d occurrence(s)", len(matches)))
				for _, match := range matches {
					previewParts = append(previewParts, fmt.Sprintf("    Line %d: %s", match.LineNumber, strings.TrimSpace(match.LineText)))
				}
				allMatches = append(allMatches, matches...)
			} else {
				previewParts = append(previewParts, fmt.Sprintf("  ✓ Would replace 1 occurrence"))
				previewParts = append(previewParts, fmt.Sprintf("    Line %d: %s", matchInfo.LineNumber, strings.TrimSpace(matchInfo.LineText)))
				allMatches = append(allMatches, *matchInfo)
			}
		} else {
			previewParts = append(previewParts, "  ✗ No matches found")
			errorMsg := fmt.Sprintf("Edit %d: string '%s' not found", i+1, edit.OldString)
			partialErrors = append(partialErrors, errorMsg)
		}
		previewParts = append(previewParts, "")
	}

	return &models.EditResult{
		Success:       len(partialErrors) == 0,
		Message:       "Dry run completed",
		PreviewDiff:   strings.Join(previewParts, "\n"),
		MatchedLines:  allMatches,
		PartialErrors: partialErrors,
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

// advancedNormalize performs advanced normalization for better matching
func advancedNormalize(s string, options *models.MatchingOptions) string {
	result := s

	// Remove excessive whitespace
	if options.IgnoreWhitespace || options.AutoNormalize {
		// Normalize all types of whitespace
		re := regexp.MustCompile(`\s+`)
		result = re.ReplaceAllString(result, " ")
		result = strings.TrimSpace(result)

		// Remove whitespace around common punctuation if AutoNormalize
		if options.AutoNormalize {
			// Remove spaces around parentheses, brackets, semicolons, etc.
			punctuationPatterns := []struct {
				pattern     string
				replacement string
			}{
				{`\s*\(\s*`, "("},
				{`\s*\)\s*`, ")"},
				{`\s*\[\s*`, "["},
				{`\s*\]\s*`, "]"},
				{`\s*;\s*`, ";"},
				{`\s*,\s*`, ","},
				{`\s*\.\s*`, "."},
				{`\s*:\s*`, ":"},
			}

			for _, p := range punctuationPatterns {
				re := regexp.MustCompile(p.pattern)
				result = re.ReplaceAllString(result, p.replacement)
			}
		}
	}

	// Case normalization
	if options.CaseInsensitive {
		result = strings.ToLower(result)
	}

	return result
}

// calculateSimilarity calculates similarity between two strings using Levenshtein distance
func calculateSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	// Use simple word-based similarity for performance
	words1 := strings.Fields(s1)
	words2 := strings.Fields(s2)

	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	matchedWords := 0
	for _, w1 := range words1 {
		for _, w2 := range words2 {
			if strings.Contains(strings.ToLower(w1), strings.ToLower(w2)) ||
			   strings.Contains(strings.ToLower(w2), strings.ToLower(w1)) {
				matchedWords++
				break
			}
		}
	}

	// Return similarity as ratio of matched words
	maxWords := len(words1)
	if len(words2) > maxWords {
		maxWords = len(words2)
	}

	return float64(matchedWords) / float64(maxWords)
}

// chunkString breaks a large string into smaller chunks for better matching
func chunkString(s string, maxSize int) []string {
	if len(s) <= maxSize {
		return []string{s}
	}

	var chunks []string
	lines := strings.Split(s, "\n")

	currentChunk := ""
	for _, line := range lines {
		// If adding this line would exceed maxSize, save current chunk and start new one
		if len(currentChunk)+len(line)+1 > maxSize && currentChunk != "" {
			chunks = append(chunks, strings.TrimSpace(currentChunk))
			currentChunk = line
		} else {
			if currentChunk != "" {
				currentChunk += "\n" + line
			} else {
				currentChunk = line
			}
		}
	}

	// Add the last chunk if not empty
	if currentChunk != "" {
		chunks = append(chunks, strings.TrimSpace(currentChunk))
	}

	return chunks
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

// advancedStringMatch performs advanced string matching with multiple strategies
func (f *FileOperations) advancedStringMatch(content, target string, options *models.MatchingOptions) (*models.MatchInfo, error) {
	lines := strings.Split(content, "\n")

	// Strategy 1: Exact match
	if matches := f.findExactMatches(content, target, lines); len(matches) > 0 {
		return &matches[0], nil
	}

	// Strategy 2: Regex matching
	if options.UseRegex {
		if matches := f.findRegexMatches(content, target, lines); len(matches) > 0 {
			return &matches[0], nil
		}
	}

	// Strategy 3: Auto-chunking for large strings
	if options.AutoChunk && len(target) > options.MaxChunkSize {
		if matches := f.findChunkedMatches(content, target, lines, options); len(matches) > 0 {
			return &matches[0], nil
		}
	}

	// Strategy 4: Normalized matching (more tolerant)
	if options.AutoNormalize {
		if matches := f.findNormalizedMatches(content, target, lines, options); len(matches) > 0 {
			return &matches[0], nil
		}
	}

	// Strategy 5: Enhanced fuzzy matching
	if options.FuzzyMatch {
		if matches := f.findEnhancedFuzzyMatches(content, target, lines, options); len(matches) > 0 {
			return &matches[0], nil
		}
	}

	return nil, fmt.Errorf("no matches found")
}

// findExactMatches finds exact string matches
func (f *FileOperations) findExactMatches(content, target string, lines []string) []models.MatchInfo {
	var matches []models.MatchInfo

	for i, line := range lines {
		if strings.Contains(line, target) {
			context := f.getLineContext(lines, i, 2)
			matches = append(matches, models.MatchInfo{
				LineNumber: i + 1,
				LineText:   line,
				MatchText:  target,
				Context:    context,
			})
		}
	}

	return matches
}

// findRegexMatches finds matches using regex
func (f *FileOperations) findRegexMatches(content, pattern string, lines []string) []models.MatchInfo {
	var matches []models.MatchInfo

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return matches
	}

	for i, line := range lines {
		if match := regex.FindString(line); match != "" {
			context := f.getLineContext(lines, i, 2)
			matches = append(matches, models.MatchInfo{
				LineNumber: i + 1,
				LineText:   line,
				MatchText:  match,
				Context:    context,
			})
		}
	}

	return matches
}

// findFuzzyMatches finds matches using fuzzy logic
func (f *FileOperations) findFuzzyMatches(content, target string, lines []string, options *models.MatchingOptions) []models.MatchInfo {
	var matches []models.MatchInfo

	normalizedTarget := f.normalizeForMatching(target, options)

	for i, line := range lines {
		normalizedLine := f.normalizeForMatching(line, options)

		// Try substring match
		if strings.Contains(normalizedLine, normalizedTarget) {
			context := f.getLineContext(lines, i, 2)
			matches = append(matches, models.MatchInfo{
				LineNumber: i + 1,
				LineText:   line,
				MatchText:  f.extractMatchFromLine(line, target),
				Context:    context,
			})
			continue
		}

		// Try word-based matching
		threshold := options.SimilarityThreshold
		if threshold == 0 {
			threshold = 0.7 // default
		}
		if f.fuzzyWordMatch(normalizedLine, normalizedTarget, threshold) {
			context := f.getLineContext(lines, i, 2)
			matches = append(matches, models.MatchInfo{
				LineNumber: i + 1,
				LineText:   line,
				MatchText:  f.extractMatchFromLine(line, target),
				Context:    context,
			})
		}
	}

	return matches
}

// findNormalizedMatches finds matches using advanced normalization
func (f *FileOperations) findNormalizedMatches(content, target string, lines []string, options *models.MatchingOptions) []models.MatchInfo {
	var matches []models.MatchInfo

	normalizedTarget := advancedNormalize(target, options)

	for i, line := range lines {
		normalizedLine := advancedNormalize(line, options)

		if strings.Contains(normalizedLine, normalizedTarget) {
			context := f.getLineContext(lines, i, 2)
			matches = append(matches, models.MatchInfo{
				LineNumber: i + 1,
				LineText:   line,
				MatchText:  f.extractBestMatch(line, target, options),
				Context:    context,
			})
		}
	}

	return matches
}

// findChunkedMatches finds matches by breaking large strings into chunks
func (f *FileOperations) findChunkedMatches(content, target string, lines []string, options *models.MatchingOptions) []models.MatchInfo {
	var matches []models.MatchInfo

	// Break target into chunks
	chunks := chunkString(target, options.MaxChunkSize)

	for _, chunk := range chunks {
		chunkMatches := f.findNormalizedMatches(content, chunk, lines, options)
		if len(chunkMatches) > 0 {
			// Found a chunk, try to find the full match around this area
			for _, chunkMatch := range chunkMatches {
				expandedMatch := f.expandMatch(lines, chunkMatch.LineNumber-1, target, options)
				if expandedMatch != nil {
					matches = append(matches, *expandedMatch)
					break // Found it, no need to continue with other chunks
				}
			}
			if len(matches) > 0 {
				break
			}
		}
	}

	return matches
}

// findEnhancedFuzzyMatches finds matches using enhanced fuzzy matching with similarity threshold
func (f *FileOperations) findEnhancedFuzzyMatches(content, target string, lines []string, options *models.MatchingOptions) []models.MatchInfo {
	var matches []models.MatchInfo

	normalizedTarget := f.normalizeForMatching(target, options)
	threshold := options.SimilarityThreshold
	if threshold == 0 {
		threshold = 0.6 // Lower threshold for enhanced matching
	}

	// Try line-by-line matching first
	for i, line := range lines {
		normalizedLine := f.normalizeForMatching(line, options)
		similarity := calculateSimilarity(normalizedLine, normalizedTarget)

		if similarity >= threshold {
			context := f.getLineContext(lines, i, 2)
			matches = append(matches, models.MatchInfo{
				LineNumber: i + 1,
				LineText:   line,
				MatchText:  f.extractBestMatch(line, target, options),
				Context:    context,
			})
		}
	}

	// If no line matches, try multi-line blocks
	if len(matches) == 0 {
		matches = f.findMultiLineMatches(content, target, lines, options)
	}

	return matches
}

// findMultiLineMatches finds matches across multiple lines
func (f *FileOperations) findMultiLineMatches(content, target string, lines []string, options *models.MatchingOptions) []models.MatchInfo {
	var matches []models.MatchInfo

	// Try blocks of 2-5 lines
	for blockSize := 2; blockSize <= 5 && len(matches) == 0; blockSize++ {
		for i := 0; i <= len(lines)-blockSize; i++ {
			block := strings.Join(lines[i:i+blockSize], "\n")
			normalizedBlock := f.normalizeForMatching(block, options)
			normalizedTarget := f.normalizeForMatching(target, options)

			similarity := calculateSimilarity(normalizedBlock, normalizedTarget)
			threshold := options.SimilarityThreshold
			if threshold == 0 {
				threshold = 0.5 // Even lower threshold for multi-line
			}

			if similarity >= threshold {
				context := f.getLineContext(lines, i+blockSize/2, 3)
				matches = append(matches, models.MatchInfo{
					LineNumber: i + 1,
					LineText:   lines[i], // Show first line of the block
					MatchText:  f.extractBestMatch(block, target, options),
					Context:    context,
				})
			}
		}
	}

	return matches
}

// expandMatch tries to expand a match around a found chunk
func (f *FileOperations) expandMatch(lines []string, startLine int, target string, options *models.MatchingOptions) *models.MatchInfo {
	if startLine < 0 || startLine >= len(lines) {
		return nil
	}

	// Try expanding the match in both directions
	for radius := 1; radius <= 10; radius++ {
		start := startLine - radius
		end := startLine + radius + 1

		if start < 0 {
			start = 0
		}
		if end > len(lines) {
			end = len(lines)
		}

		block := strings.Join(lines[start:end], "\n")
		normalizedBlock := f.normalizeForMatching(block, options)
		normalizedTarget := f.normalizeForMatching(target, options)

		similarity := calculateSimilarity(normalizedBlock, normalizedTarget)
		if similarity >= 0.7 { // High threshold for expanded matches
			context := f.getLineContext(lines, startLine, 3)
			return &models.MatchInfo{
				LineNumber: start + 1,
				LineText:   lines[startLine],
				MatchText:  f.extractBestMatch(block, target, options),
				Context:    context,
			}
		}
	}

	return nil
}

// extractBestMatch extracts the best matching part considering options
func (f *FileOperations) extractBestMatch(content, target string, options *models.MatchingOptions) string {
	// Try exact match first
	if strings.Contains(content, target) {
		return target
	}

	// Try normalized match
	if options.AutoNormalize {
		normalizedContent := advancedNormalize(content, options)
		normalizedTarget := advancedNormalize(target, options)

		if strings.Contains(normalizedContent, normalizedTarget) {
			// Find the original text that corresponds to the normalized match
			return f.findOriginalMatch(content, target, options)
		}
	}

	// Return the content trimmed if no exact match
	return strings.TrimSpace(content)
}

// findOriginalMatch finds the original text that matches the normalized version
func (f *FileOperations) findOriginalMatch(content, target string, options *models.MatchingOptions) string {
	lines := strings.Split(content, "\n")
	targetWords := strings.Fields(advancedNormalize(target, options))

	bestMatch := ""
	bestScore := 0.0

	for _, line := range lines {
		normalizedLine := advancedNormalize(line, options)
		lineWords := strings.Fields(normalizedLine)

		score := 0.0
		for _, targetWord := range targetWords {
			for _, lineWord := range lineWords {
				if strings.Contains(lineWord, targetWord) || strings.Contains(targetWord, lineWord) {
					score += 1.0
				}
			}
		}

		if len(targetWords) > 0 {
			score = score / float64(len(targetWords))
		}

		if score > bestScore {
			bestScore = score
			bestMatch = strings.TrimSpace(line)
		}
	}

	if bestMatch != "" {
		return bestMatch
	}

	return strings.TrimSpace(content)
}

// normalizeForMatching normalizes strings for fuzzy matching
func (f *FileOperations) normalizeForMatching(s string, options *models.MatchingOptions) string {
	return advancedNormalize(s, options)
}

// fuzzyWordMatch performs word-based fuzzy matching
func (f *FileOperations) fuzzyWordMatch(line, target string, threshold float64) bool {
	similarity := calculateSimilarity(line, target)
	return similarity >= threshold
}

// extractMatchFromLine extracts the best matching part from a line
func (f *FileOperations) extractMatchFromLine(line, target string) string {
	// Try exact match first
	if strings.Contains(line, target) {
		return target
	}

	// Return the line trimmed if no exact match
	return strings.TrimSpace(line)
}

// getLineContext returns surrounding lines for context
func (f *FileOperations) getLineContext(lines []string, lineIndex, contextSize int) []string {
	start := lineIndex - contextSize
	if start < 0 {
		start = 0
	}

	end := lineIndex + contextSize + 1
	if end > len(lines) {
		end = len(lines)
	}

	context := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		prefix := "  "
		if i == lineIndex {
			prefix = "> "
		}
		context = append(context, fmt.Sprintf("%s%d: %s", prefix, i+1, lines[i]))
	}

	return context
}

// generateDiffPreview generates a detailed diff-like preview of changes
func (f *FileOperations) generateDiffPreview(original, modified string, matchInfo *models.MatchInfo) string {
	var diff strings.Builder

	diff.WriteString("==== PREVIEW MODE - Detailed Change Analysis ====\n\n")

	if matchInfo != nil {
		diff.WriteString(fmt.Sprintf("📍 Match Location: Line %d\n", matchInfo.LineNumber))
		diff.WriteString(fmt.Sprintf("🔍 Match Type: %s\n", f.getMatchType(matchInfo)))
		diff.WriteString(fmt.Sprintf("📏 Original Length: %d characters\n", len(matchInfo.MatchText)))
		diff.WriteString(fmt.Sprintf("📏 New Length: %d characters\n", len(modified)))

		// Calculate change statistics
		if len(modified) > len(matchInfo.MatchText) {
			diff.WriteString(fmt.Sprintf("📈 Change: +%d characters\n", len(modified)-len(matchInfo.MatchText)))
		} else if len(modified) < len(matchInfo.MatchText) {
			diff.WriteString(fmt.Sprintf("📉 Change: -%d characters\n", len(matchInfo.MatchText)-len(modified)))
		} else {
			diff.WriteString("🔄 Change: Same length (substitution)\n")
		}

		diff.WriteString("\n")

		// Show context with line numbers
		if len(matchInfo.Context) > 0 {
			diff.WriteString("📋 Context (showing surrounding lines):\n")
			diff.WriteString("────────────────────────────────────────\n")
			for _, contextLine := range matchInfo.Context {
				diff.WriteString(contextLine + "\n")
			}
			diff.WriteString("────────────────────────────────────────\n\n")
		}

		// Show detailed change preview
		if matchInfo.MatchText != "" {
			diff.WriteString("🔄 Exact Changes:\n")
			diff.WriteString("────────────────\n")

			// Show original with highlighting
			diff.WriteString("BEFORE:\n")
			if len(matchInfo.MatchText) > 100 {
				diff.WriteString(fmt.Sprintf("  (showing first 100 chars) %s...\n",
					f.sanitizeForPreview(matchInfo.MatchText[:100])))
			} else {
				diff.WriteString(fmt.Sprintf("  %s\n", f.sanitizeForPreview(matchInfo.MatchText)))
			}

			diff.WriteString("\nAFTER:\n")
			if len(modified) > 100 {
				diff.WriteString(fmt.Sprintf("  (showing first 100 chars) %s...\n",
					f.sanitizeForPreview(modified[:100])))
			} else {
				diff.WriteString(fmt.Sprintf("  %s\n", f.sanitizeForPreview(modified)))
			}

			// Show word-level diff if strings are reasonable length
			if len(matchInfo.MatchText) < 500 && len(modified) < 500 {
				diff.WriteString("\n📝 Word-level Changes:\n")
				wordDiff := f.generateWordDiff(matchInfo.MatchText, modified)
				diff.WriteString(wordDiff)
			}
		}

		// Safety checks
		diff.WriteString("\n🛡️  Safety Checks:\n")
		diff.WriteString("─────────────────\n")

		if strings.Contains(matchInfo.MatchText, "procedure ") || strings.Contains(matchInfo.MatchText, "function ") {
			diff.WriteString("⚠️  Code block detected - this appears to be a procedure/function\n")
		}
		if strings.Count(matchInfo.MatchText, "\n") > 10 {
			diff.WriteString("⚠️  Large text block - this affects multiple lines\n")
		}
		if len(matchInfo.MatchText) > 1000 {
			diff.WriteString("⚠️  Large change - consider breaking into smaller edits\n")
		}
		if matchInfo.MatchText == modified {
			diff.WriteString("❌ No actual change - old and new text are identical\n")
		} else {
			diff.WriteString("✅ Valid change detected\n")
		}
	}

	diff.WriteString("\n💡 To apply these changes, run the same command without --preview\n")

	return diff.String()
}

// getMatchType determines what type of match was found
func (f *FileOperations) getMatchType(matchInfo *models.MatchInfo) string {
	if strings.Contains(matchInfo.MatchText, matchInfo.LineText) {
		return "Exact match"
	}
	if len(matchInfo.MatchText) < len(matchInfo.LineText) {
		return "Partial line match"
	}
	return "Fuzzy/Normalized match"
}

// sanitizeForPreview sanitizes text for safe preview display
func (f *FileOperations) sanitizeForPreview(text string) string {
	// Replace tabs with visible characters for better preview
	text = strings.ReplaceAll(text, "\t", "→")
	// Trim excessive whitespace but show it exists
	if strings.TrimSpace(text) != text {
		if strings.HasPrefix(text, " ") || strings.HasPrefix(text, "\t") {
			text = "␣" + strings.TrimLeft(text, " \t")
		}
		if strings.HasSuffix(text, " ") || strings.HasSuffix(text, "\t") {
			text = strings.TrimRight(text, " \t") + "␣"
		}
	}
	return text
}

// generateWordDiff generates a word-level diff between two strings
func (f *FileOperations) generateWordDiff(old, new string) string {
	oldWords := strings.Fields(old)
	newWords := strings.Fields(new)

	var diff strings.Builder

	// Simple word-by-word comparison
	maxLen := len(oldWords)
	if len(newWords) > maxLen {
		maxLen = len(newWords)
	}

	for i := 0; i < maxLen; i++ {
		oldWord := ""
		newWord := ""

		if i < len(oldWords) {
			oldWord = oldWords[i]
		}
		if i < len(newWords) {
			newWord = newWords[i]
		}

		if oldWord == newWord {
			if oldWord != "" {
				diff.WriteString(fmt.Sprintf("  %s\n", oldWord))
			}
		} else {
			if oldWord != "" {
				diff.WriteString(fmt.Sprintf("- %s\n", oldWord))
			}
			if newWord != "" {
				diff.WriteString(fmt.Sprintf("+ %s\n", newWord))
			}
		}
	}

	return diff.String()
}

// findStringMatches returns all possible matches with context (enhanced version)
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

// findAllMatches finds all matches using advanced matching strategies
func (f *FileOperations) findAllMatches(content, target string, options *models.MatchingOptions) []models.MatchInfo {
	lines := strings.Split(content, "\n")
	var allMatches []models.MatchInfo

	// Strategy 1: Exact matches first
	exactMatches := f.findExactMatches(content, target, lines)
	allMatches = append(allMatches, exactMatches...)

	// Strategy 2: If no exact matches and regex is enabled, try regex
	if len(allMatches) == 0 && options.UseRegex {
		regexMatches := f.findRegexMatches(content, target, lines)
		allMatches = append(allMatches, regexMatches...)
	}

	// Strategy 3: Auto-chunking for large strings
	if len(allMatches) == 0 && options.AutoChunk && len(target) > options.MaxChunkSize {
		chunkMatches := f.findChunkedMatches(content, target, lines, options)
		allMatches = append(allMatches, chunkMatches...)
	}

	// Strategy 4: Normalized matching (more tolerant)
	if len(allMatches) == 0 && options.AutoNormalize {
		normalizedMatches := f.findNormalizedMatches(content, target, lines, options)
		allMatches = append(allMatches, normalizedMatches...)
	}

	// Strategy 5: Enhanced fuzzy matching
	if len(allMatches) == 0 && options.FuzzyMatch {
		fuzzyMatches := f.findEnhancedFuzzyMatches(content, target, lines, options)
		allMatches = append(allMatches, fuzzyMatches...)
	}

	// Strategy 6: If still no matches, try the original fuzzy matching as fallback
	if len(allMatches) == 0 && options.FuzzyMatch {
		originalFuzzyMatches := f.findFuzzyMatches(content, target, lines, options)
		allMatches = append(allMatches, originalFuzzyMatches...)
	}

	return allMatches
}

// replaceAllMatches replaces all matched strings
func (f *FileOperations) replaceAllMatches(content string, matches []models.MatchInfo, newString string) string {
	lines := strings.Split(content, "\n")

	// Process matches in reverse order to maintain line numbers
	for i := len(matches) - 1; i >= 0; i-- {
		match := matches[i]
		lineIndex := match.LineNumber - 1
		if lineIndex >= 0 && lineIndex < len(lines) {
			lines[lineIndex] = strings.Replace(lines[lineIndex], match.MatchText, newString, 1)
		}
	}

	return strings.Join(lines, "\n")
}

// replaceSingleMatch replaces a single matched string
func (f *FileOperations) replaceSingleMatch(content string, match *models.MatchInfo, newString string) string {
	lines := strings.Split(content, "\n")
	lineIndex := match.LineNumber - 1

	if lineIndex >= 0 && lineIndex < len(lines) {
		lines[lineIndex] = strings.Replace(lines[lineIndex], match.MatchText, newString, 1)
	}

	return strings.Join(lines, "\n")
}

// handleNoMatchesFound handles the case when no matches are found
func (f *FileOperations) handleNoMatchesFound(content, target string) (*models.EditResult, error) {
	matches := findStringMatches(content, target)
	errorMsg := fmt.Sprintf("String '%s' not found in file", target)
	if len(matches) > 0 {
		errorMsg += "\nSimilar matches found:\n" + strings.Join(matches, "\n")
	}

	// Try fuzzy matching for suggestions
	if lineIndex, matchedLine := findBestMatch(content, target); lineIndex != -1 {
		errorMsg += fmt.Sprintf("\nBest fuzzy match found at line %d: %q", lineIndex+1, matchedLine)
	}

	return &models.EditResult{
		Success: false,
		Message: errorMsg,
		Error:   fmt.Errorf("old string not found"),
	}, nil
}

// handleMultipleMatches handles the case when multiple matches are found but replaceAll is false
func (f *FileOperations) handleMultipleMatches(content, target string, count int) (*models.EditResult, error) {
	lines := strings.Split(content, "\n")
	var matchLines []string
	var matchInfos []models.MatchInfo

	for i, line := range lines {
		if strings.Contains(line, target) {
			matchLines = append(matchLines, fmt.Sprintf("Line %d: %q", i+1, strings.TrimSpace(line)))
			context := f.getLineContext(lines, i, 1)
			matchInfos = append(matchInfos, models.MatchInfo{
				LineNumber: i + 1,
				LineText:   line,
				MatchText:  target,
				Context:    context,
			})
		}
	}

	errorMsg := fmt.Sprintf("String '%s' appears %d times in file, use --replace-all flag", target, count)
	if len(matchLines) > 0 {
		errorMsg += "\nMatches found at:\n" + strings.Join(matchLines, "\n")
	}

	return &models.EditResult{
		Success:      false,
		Message:      errorMsg,
		Error:        fmt.Errorf("string not unique"),
		MatchedLines: matchInfos,
	}, nil
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