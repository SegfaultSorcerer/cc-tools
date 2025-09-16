package models

// EditOperation represents a single edit operation
type EditOperation struct {
	OldString  string `json:"old_string"`
	NewString  string `json:"new_string"`
	ReplaceAll bool   `json:"replace_all,omitempty"`
	UseRegex   bool   `json:"use_regex,omitempty"`
	FuzzyMatch bool   `json:"fuzzy_match,omitempty"`
}

// MultiEditRequest represents multiple edit operations on a single file
type MultiEditRequest struct {
	FilePath        string          `json:"file_path"`
	Edits           []EditOperation `json:"edits"`
	ContinueOnError bool            `json:"continue_on_error,omitempty"`
	DryRun          bool            `json:"dry_run,omitempty"`
}

// FileInfo holds information about a file including its encoding
type FileInfo struct {
	Path     string
	Encoding string
	Content  []byte
}

// EditResult represents the result of an edit operation
type EditResult struct {
	Success       bool
	Message       string
	Error         error
	PreviewDiff   string // Diff preview of changes
	MatchedLines  []MatchInfo // Information about matched lines
	PartialErrors []string // Errors for individual operations in multiedit
}

// MatchInfo represents information about a matched string
type MatchInfo struct {
	LineNumber int
	LineText   string
	MatchText  string
	Context    []string // Surrounding lines for context
	Uniqueness float64 // Uniqueness score for disambiguation
}

// MatchingOptions represents options for string matching
type MatchingOptions struct {
	UseRegex         bool
	FuzzyMatch       bool
	IgnoreWhitespace bool
	CaseInsensitive  bool
	AutoNormalize    bool  // Automatically normalize whitespace and formatting
	SimilarityThreshold float64 // Threshold for fuzzy matching (0.0-1.0)
	AutoChunk        bool  // Automatically break large strings into smaller chunks
	MaxChunkSize     int   // Maximum size for chunks when AutoChunk is enabled
	SmartCode        bool  // Enable smart code understanding for better block matching
	AggressiveFuzzy  bool  // Enable more aggressive fuzzy matching for irregular formatting
	SmartSuggestions bool  // Enable intelligent suggestions when exact match fails
	CodeLanguage     string // Programming language hint (auto-detected if empty)
	DebugMode        bool  // Enable debug mode for troubleshooting matching issues
}

// CopyOperation represents a file copy operation
type CopyOperation struct {
	SourcePath      string `json:"source_path"`
	DestinationPath string `json:"destination_path"`
	PreserveMode    bool   `json:"preserve_mode,omitempty"`
	Overwrite       bool   `json:"overwrite,omitempty"`
}

// MoveOperation represents a file move operation
type MoveOperation struct {
	SourcePath      string `json:"source_path"`
	DestinationPath string `json:"destination_path"`
	Overwrite       bool   `json:"overwrite,omitempty"`
}

// DeleteOperation represents a file delete operation
type DeleteOperation struct {
	FilePath    string `json:"file_path"`
	CreateBackup bool   `json:"create_backup,omitempty"`
	BackupPath   string `json:"backup_path,omitempty"`
}

// FileOperationResult represents the result of file operations
type FileOperationResult struct {
	Success     bool
	Message     string
	Error       error
	BackupPath  string // Path to backup file if created
	SourceInfo  *FileInfo // Information about source file
	TargetInfo  *FileInfo // Information about target file (for copy/move)
}

// DirectoryOperation represents a directory creation operation
type DirectoryOperation struct {
	Path        string `json:"path"`
	CreateParents bool   `json:"create_parents,omitempty"`
	Mode        int    `json:"mode,omitempty"`
}

// DirectoryCopyOperation represents a directory copy operation
type DirectoryCopyOperation struct {
	SourcePath      string `json:"source_path"`
	DestinationPath string `json:"destination_path"`
	PreserveAll     bool   `json:"preserve_all,omitempty"`
	Overwrite       bool   `json:"overwrite,omitempty"`
	SkipExisting    bool   `json:"skip_existing,omitempty"`
}

// DirectoryMoveOperation represents a directory move operation
type DirectoryMoveOperation struct {
	SourcePath      string `json:"source_path"`
	DestinationPath string `json:"destination_path"`
	Overwrite       bool   `json:"overwrite,omitempty"`
}

// DirectoryDeleteOperation represents a directory delete operation
type DirectoryDeleteOperation struct {
	Path         string `json:"path"`
	Recursive    bool   `json:"recursive,omitempty"`
	CreateBackup bool   `json:"create_backup,omitempty"`
	BackupPath   string `json:"backup_path,omitempty"`
}

// DirectoryListOperation represents a directory listing operation
type DirectoryListOperation struct {
	Path           string `json:"path"`
	Recursive      bool   `json:"recursive,omitempty"`
	ShowEncoding   bool   `json:"show_encoding,omitempty"`
	Filter         string `json:"filter,omitempty"`
	ShowHidden     bool   `json:"show_hidden,omitempty"`
}

// DirectoryInfo holds information about a directory
type DirectoryInfo struct {
	Path          string
	TotalFiles    int
	TotalSize     int64
	Encodings     map[string]int // encoding -> count
	FileTypes     map[string]int // extension -> count
	Subdirectories []string
}

// FileEntry represents a file in directory listing
type FileEntry struct {
	Path     string
	Name     string
	Size     int64
	IsDir    bool
	Mode     string
	Encoding string
}

// DirectoryOperationResult represents the result of directory operations
type DirectoryOperationResult struct {
	Success        bool
	Message        string
	Error          error
	BackupPath     string
	ProcessedFiles int
	ProcessedDirs  int
	TotalSize      int64
	SourceInfo     *DirectoryInfo
	TargetInfo     *DirectoryInfo
	FileList       []FileEntry
}

// ProgressInfo represents progress information for long operations
type ProgressInfo struct {
	CurrentFile   string
	FilesProcessed int
	TotalFiles    int
	BytesProcessed int64
	TotalBytes    int64
	StartTime     int64
}