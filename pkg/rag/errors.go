// Package rag provides error types for RAG operations
package rag

import "fmt"

// Error types for RAG operations

// ErrInvalidURL represents an invalid URL error
type ErrInvalidURL string

func (e ErrInvalidURL) Error() string {
	return fmt.Sprintf("invalid URL: %s", string(e))
}

// ErrInvalidPath represents an invalid file path error
type ErrInvalidPath string

func (e ErrInvalidPath) Error() string {
	return fmt.Sprintf("invalid path: %s", string(e))
}

// ErrDownloadFailed represents a download failure error
type ErrDownloadFailed struct {
	URL    string
	Reason string
}

func (e ErrDownloadFailed) Error() string {
	return fmt.Sprintf("download failed for %s: %s", e.URL, e.Reason)
}

// ErrCacheError represents a cache operation error
type ErrCacheError struct {
	Operation string
	Reason    string
}

func (e ErrCacheError) Error() string {
	return fmt.Sprintf("cache error during %s: %s", e.Operation, e.Reason)
}

// ErrInvalidConfig represents an invalid configuration error
type ErrInvalidConfig string

func (e ErrInvalidConfig) Error() string {
	return fmt.Sprintf("invalid configuration: %s", string(e))
}

// ErrFileTooLarge represents a file size limit exceeded error
type ErrFileTooLarge struct {
	Size    int64
	MaxSize int64
}

func (e ErrFileTooLarge) Error() string {
	return fmt.Sprintf("file size %d bytes exceeds maximum %d bytes", e.Size, e.MaxSize)
}

// ErrNotTextFile represents an error when a file is not text-based
type ErrNotTextFile string

func (e ErrNotTextFile) Error() string {
	return fmt.Sprintf("not a text file: %s", string(e))
}

// ErrPathTraversal represents a path traversal attempt error
type ErrPathTraversal string

func (e ErrPathTraversal) Error() string {
	return fmt.Sprintf("path traversal detected: %s", string(e))
}
