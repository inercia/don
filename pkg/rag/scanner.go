// Package rag provides file and directory scanning functionality
package rag

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/inercia/don/pkg/common"
)

// FileScanner implements the Scanner interface
type FileScanner struct {
	logger *common.Logger
}

// NewScanner creates a new file scanner
func NewScanner(logger *common.Logger) *FileScanner {
	return &FileScanner{
		logger: logger,
	}
}

// ScanDirectory scans a directory for text files
func (s *FileScanner) ScanDirectory(ctx context.Context, dirPath string, recursive bool) ([]string, error) {
	// Validate path
	if err := ValidatePath(dirPath); err != nil {
		return nil, fmt.Errorf("invalid directory path: %w", err)
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Check if directory exists
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat directory: %w", err)
	}

	if !info.IsDir() {
		return nil, ErrInvalidPath(fmt.Sprintf("%s is not a directory", absPath))
	}

	var files []string

	// Walk the directory
	walkFunc := func(path string, info os.FileInfo, err error) error {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			s.logger.Warn("Error accessing path %s: %v", path, err)
			return nil // Continue walking
		}

		// Skip directories (unless we're at the root)
		if info.IsDir() {
			if path != absPath && !recursive {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		// Check if it's a text file
		if s.IsTextFile(path) {
			files = append(files, path)
			s.logger.Debug("Found text file: %s", path)
		}

		return nil
	}

	if err := filepath.Walk(absPath, walkFunc); err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	s.logger.Info("Scanned directory %s: found %d text files", dirPath, len(files))
	return files, nil
}

// ScanFile validates and returns a file path if it's a text file
func (s *FileScanner) ScanFile(filePath string) (string, error) {
	// Validate path
	if err := ValidatePath(filePath); err != nil {
		return "", fmt.Errorf("invalid file path: %w", err)
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Check if file exists
	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	if info.IsDir() {
		return "", ErrInvalidPath(fmt.Sprintf("%s is a directory, not a file", absPath))
	}

	// Check if it's a text file
	if !s.IsTextFile(absPath) {
		return "", ErrNotTextFile(absPath)
	}

	s.logger.Debug("Validated text file: %s", absPath)
	return absPath, nil
}

// IsTextFile checks if a file is text-based by extension
func (s *FileScanner) IsTextFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))

	// List of text file extensions
	textExtensions := []string{
		".txt", ".md", ".markdown",
		".json", ".yaml", ".yml",
		".xml", ".html", ".htm",
		".csv", ".tsv",
		".log",
		".rst", ".adoc", ".asciidoc",
		".tex", ".latex",
		".org",
		".conf", ".config", ".cfg",
		".ini", ".toml",
		".sh", ".bash", ".zsh",
		".py", ".js", ".ts", ".go", ".java", ".c", ".cpp", ".h", ".hpp",
		".rb", ".php", ".pl", ".lua", ".r",
		".css", ".scss", ".sass", ".less",
		".sql",
	}

	for _, textExt := range textExtensions {
		if ext == textExt {
			return true
		}
	}

	return false
}
