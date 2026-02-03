// Package rag provides cache management for downloaded documents
package rag

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	// CacheDirEnv is the environment variable for overriding the cache directory
	CacheDirEnv = "MCPSHELL_RAG_CACHE_DIR"

	// Default cache directory names
	cacheSubDir = "don/rag"

	// Cache file extensions
	metadataExt = ".meta.json"
	documentExt = ".txt"
)

// GetCacheDir returns the platform-specific cache directory for RAG documents
// Priority: MCPSHELL_RAG_CACHE_DIR env var > platform default
func GetCacheDir() (string, error) {
	// Check environment variable first
	if cacheDir := os.Getenv(CacheDirEnv); cacheDir != "" {
		return cacheDir, nil
	}

	// Get platform-specific cache directory
	var baseDir string
	switch runtime.GOOS {
	case "darwin":
		// macOS: ~/Library/Caches/don/rag
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		baseDir = filepath.Join(home, "Library", "Caches", cacheSubDir)

	case "windows":
		// Windows: %LOCALAPPDATA%\don\rag
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			return "", fmt.Errorf("LOCALAPPDATA environment variable not set")
		}
		baseDir = filepath.Join(localAppData, cacheSubDir)

	default:
		// Linux and others: $XDG_CACHE_HOME/mcpshell/rag or ~/.cache/mcpshell/rag
		xdgCache := os.Getenv("XDG_CACHE_HOME")
		if xdgCache != "" {
			baseDir = filepath.Join(xdgCache, cacheSubDir)
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to get home directory: %w", err)
			}
			baseDir = filepath.Join(home, ".cache", cacheSubDir)
		}
	}

	return baseDir, nil
}

// EnsureCacheDir creates the cache directory if it doesn't exist
func EnsureCacheDir(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory %s: %w", dir, err)
	}
	return nil
}

// GetCachedDocumentPath generates the cache file path from a URL hash
func GetCachedDocumentPath(cacheDir, url string) string {
	hash := hashURL(url)
	return filepath.Join(cacheDir, "documents", hash+documentExt)
}

// GetCachedMetadataPath generates the metadata file path from a URL hash
func GetCachedMetadataPath(cacheDir, url string) string {
	hash := hashURL(url)
	return filepath.Join(cacheDir, "documents", hash+metadataExt)
}

// hashURL generates a SHA256 hash of a URL for use as a cache key
func hashURL(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])
}

// LoadMetadata loads cache metadata from a JSON file
func LoadMetadata(path string) (*CacheMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var meta CacheMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return &meta, nil
}

// SaveMetadata saves cache metadata to a JSON file
func SaveMetadata(path string, meta *CacheMetadata) error {
	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// ValidatePath checks if a path is safe (no path traversal)
func ValidatePath(path string) error {
	// Check for path traversal attempts before cleaning
	if strings.Contains(path, "..") {
		return ErrPathTraversal(path)
	}

	// Clean the path
	cleanPath := filepath.Clean(path)

	// Check again after cleaning
	if strings.Contains(cleanPath, "..") {
		return ErrPathTraversal(path)
	}

	return nil
}

// ComputeContentHash computes the SHA256 hash of file content
func ComputeContentHash(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}
