// Package rag provides Retrieval-Augmented Generation support for Don.
// It handles document downloading, caching, and scanning for RAG functionality.
package rag

import (
	"context"
	"time"

	"github.com/inercia/don/pkg/common"
)

// SourceType represents the type of document source
type SourceType int

const (
	// SourceTypeURL represents a document from a URL
	SourceTypeURL SourceType = iota
	// SourceTypeFile represents a local file
	SourceTypeFile
	// SourceTypeDirectory represents a local directory
	SourceTypeDirectory
)

// String returns the string representation of SourceType
func (s SourceType) String() string {
	switch s {
	case SourceTypeURL:
		return "URL"
	case SourceTypeFile:
		return "File"
	case SourceTypeDirectory:
		return "Directory"
	default:
		return "Unknown"
	}
}

// DocumentSource represents a source of documents for RAG
type DocumentSource struct {
	Type     SourceType // Type of source (URL, File, Directory)
	Location string     // URL or file path
	Metadata *Metadata  // Optional metadata
}

// Metadata stores additional information about a document source
type Metadata struct {
	Description string            // Human-readable description
	Tags        []string          // Optional tags for categorization
	Properties  map[string]string // Additional properties
}

// CacheMetadata stores information about cached documents
type CacheMetadata struct {
	URL          string    `json:"url"`                     // Original URL
	DownloadedAt time.Time `json:"downloaded_at"`           // When the document was downloaded
	ContentHash  string    `json:"content_hash"`            // SHA256 hash of content
	ETag         string    `json:"etag,omitempty"`          // HTTP ETag header
	LastModified string    `json:"last_modified,omitempty"` // HTTP Last-Modified header
	ContentType  string    `json:"content_type"`            // HTTP Content-Type header
	Size         int64     `json:"size"`                    // Size in bytes
}

// DownloaderConfig holds configuration for the document downloader
type DownloaderConfig struct {
	CacheDir     string         // Cache directory path (empty uses default)
	Timeout      time.Duration  // HTTP timeout for downloads
	MaxSize      int64          // Maximum file size in bytes
	ForceRefresh bool           // Force re-download of cached documents
	Logger       *common.Logger // Logger instance
}

// Downloader interface for downloading and caching documents
type Downloader interface {
	// Download downloads a document from a URL and returns the local file path
	Download(ctx context.Context, url string) (string, error)

	// GetCachedPath returns the cached file path for a URL if it exists
	// Returns the path, whether it exists, and any error
	GetCachedPath(url string) (string, bool, error)

	// ValidateCache checks if a cached document is still fresh
	// Returns true if the cache is valid, false if it needs refresh
	ValidateCache(url string) (bool, error)

	// GetCacheDir returns the cache directory being used
	GetCacheDir() string
}

// Scanner interface for scanning local files and directories
type Scanner interface {
	// ScanDirectory scans a directory for text files
	// Returns a list of file paths
	ScanDirectory(ctx context.Context, dirPath string, recursive bool) ([]string, error)

	// ScanFile validates and returns a file path if it's a text file
	ScanFile(filePath string) (string, error)

	// IsTextFile checks if a file is text-based
	IsTextFile(path string) bool
}

// DefaultConfig returns a default DownloaderConfig
func DefaultConfig(logger *common.Logger) DownloaderConfig {
	return DownloaderConfig{
		CacheDir:     "", // Empty string uses platform default
		Timeout:      30 * time.Second,
		MaxSize:      10 * 1024 * 1024, // 10MB
		ForceRefresh: false,
		Logger:       logger,
	}
}

// ValidateConfig validates the downloader configuration
func ValidateConfig(config DownloaderConfig) error {
	if config.Logger == nil {
		return ErrInvalidConfig("logger is required")
	}
	if config.Timeout <= 0 {
		return ErrInvalidConfig("timeout must be positive")
	}
	if config.MaxSize <= 0 {
		return ErrInvalidConfig("max size must be positive")
	}
	return nil
}
