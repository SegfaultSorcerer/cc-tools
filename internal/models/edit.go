package models

// EditOperation represents a single edit operation
type EditOperation struct {
	OldString  string `json:"old_string"`
	NewString  string `json:"new_string"`
	ReplaceAll bool   `json:"replace_all,omitempty"`
}

// MultiEditRequest represents multiple edit operations on a single file
type MultiEditRequest struct {
	FilePath string          `json:"file_path"`
	Edits    []EditOperation `json:"edits"`
}

// FileInfo holds information about a file including its encoding
type FileInfo struct {
	Path     string
	Encoding string
	Content  []byte
}

// EditResult represents the result of an edit operation
type EditResult struct {
	Success bool
	Message string
	Error   error
}