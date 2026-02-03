// Package rag provides document downloading functionality
package rag

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/inercia/don/pkg/common"
)

// DocumentDownloader implements the Downloader interface
type DocumentDownloader struct {
	config     DownloaderConfig
	cacheDir   string
	logger     *common.Logger
	httpClient *http.Client
}

// NewDownloader creates a new document downloader
func NewDownloader(config DownloaderConfig) (*DocumentDownloader, error) {
	// Validate configuration
	if err := ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Determine cache directory
	cacheDir := config.CacheDir
	if cacheDir == "" {
		var err error
		cacheDir, err = GetCacheDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get cache directory: %w", err)
		}
	}

	// Ensure cache directory exists
	if err := EnsureCacheDir(cacheDir); err != nil {
		return nil, fmt.Errorf("failed to ensure cache directory: %w", err)
	}

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	config.Logger.Debug("Initialized RAG downloader with cache directory: %s", cacheDir)

	return &DocumentDownloader{
		config:     config,
		cacheDir:   cacheDir,
		logger:     config.Logger,
		httpClient: httpClient,
	}, nil
}

// GetCacheDir returns the cache directory being used
func (d *DocumentDownloader) GetCacheDir() string {
	return d.cacheDir
}

// Download downloads a document from a URL and returns the local file path
func (d *DocumentDownloader) Download(ctx context.Context, urlStr string) (string, error) {
	// Validate URL
	if err := d.validateURL(urlStr); err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Check if we have a valid cached version
	cachedPath, exists, err := d.GetCachedPath(urlStr)
	if err != nil {
		d.logger.Warn("Failed to check cache for %s: %v", urlStr, err)
	}

	if exists && !d.config.ForceRefresh {
		// Validate cache freshness
		valid, err := d.ValidateCache(urlStr)
		if err != nil {
			d.logger.Warn("Failed to validate cache for %s: %v", urlStr, err)
		} else if valid {
			d.logger.Debug("Using cached document for %s", urlStr)
			return cachedPath, nil
		}
	}

	// Download the document
	d.logger.Info("Downloading document from %s", urlStr)
	content, meta, err := d.downloadWithValidation(ctx, urlStr)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}

	// Save to cache
	docPath := GetCachedDocumentPath(d.cacheDir, urlStr)
	metaPath := GetCachedMetadataPath(d.cacheDir, urlStr)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(docPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Write document content
	if err := os.WriteFile(docPath, content, 0644); err != nil {
		return "", fmt.Errorf("failed to write document to cache: %w", err)
	}

	// Write metadata
	if err := SaveMetadata(metaPath, meta); err != nil {
		d.logger.Warn("Failed to save metadata for %s: %v", urlStr, err)
	}

	d.logger.Info("Downloaded and cached document from %s (%d bytes)", urlStr, len(content))
	return docPath, nil
}

// GetCachedPath returns the cached file path for a URL if it exists
func (d *DocumentDownloader) GetCachedPath(urlStr string) (string, bool, error) {
	docPath := GetCachedDocumentPath(d.cacheDir, urlStr)

	// Check if file exists
	if _, err := os.Stat(docPath); err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("failed to stat cached file: %w", err)
	}

	return docPath, true, nil
}

// ValidateCache checks if a cached document is still fresh
func (d *DocumentDownloader) ValidateCache(urlStr string) (bool, error) {
	metaPath := GetCachedMetadataPath(d.cacheDir, urlStr)

	// Load metadata
	meta, err := LoadMetadata(metaPath)
	if err != nil {
		return false, fmt.Errorf("failed to load metadata: %w", err)
	}

	// Check if we have ETag or Last-Modified
	if meta.ETag == "" && meta.LastModified == "" {
		// No validation headers, assume cache is valid
		return true, nil
	}

	// Make HEAD request to check freshness
	return d.checkCacheFreshness(urlStr, meta)
}

// validateURL validates that a URL is safe to download from
func (d *DocumentDownloader) validateURL(urlStr string) error {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ErrInvalidURL(fmt.Sprintf("failed to parse: %v", err))
	}

	// Only allow http and https schemes
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ErrInvalidURL(fmt.Sprintf("unsupported scheme: %s", parsedURL.Scheme))
	}

	// Ensure host is present
	if parsedURL.Host == "" {
		return ErrInvalidURL("missing host")
	}

	return nil
}

// downloadWithValidation downloads content from a URL with validation
func (d *DocumentDownloader) downloadWithValidation(ctx context.Context, urlStr string) ([]byte, *CacheMetadata, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, nil, ErrDownloadFailed{URL: urlStr, Reason: err.Error()}
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, nil, ErrDownloadFailed{
			URL:    urlStr,
			Reason: fmt.Sprintf("HTTP %d", resp.StatusCode),
		}
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if !isTextContentType(contentType) {
		return nil, nil, ErrNotTextFile(fmt.Sprintf("content type: %s", contentType))
	}

	// Check content length
	if resp.ContentLength > d.config.MaxSize {
		return nil, nil, ErrFileTooLarge{
			Size:    resp.ContentLength,
			MaxSize: d.config.MaxSize,
		}
	}

	// Read content with size limit
	limitedReader := io.LimitReader(resp.Body, d.config.MaxSize+1)
	content, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check if we exceeded the size limit
	if int64(len(content)) > d.config.MaxSize {
		return nil, nil, ErrFileTooLarge{
			Size:    int64(len(content)),
			MaxSize: d.config.MaxSize,
		}
	}

	// Create metadata
	meta := &CacheMetadata{
		URL:          urlStr,
		DownloadedAt: time.Now(),
		ContentHash:  ComputeContentHash(content),
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
		ContentType:  contentType,
		Size:         int64(len(content)),
	}

	return content, meta, nil
}

// checkCacheFreshness makes a HEAD request to check if cached content is still fresh
func (d *DocumentDownloader) checkCacheFreshness(urlStr string, meta *CacheMetadata) (bool, error) {
	req, err := http.NewRequest("HEAD", urlStr, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create HEAD request: %w", err)
	}

	// Add conditional headers
	if meta.ETag != "" {
		req.Header.Set("If-None-Match", meta.ETag)
	}
	if meta.LastModified != "" {
		req.Header.Set("If-Modified-Since", meta.LastModified)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		// If HEAD request fails, assume cache is valid
		d.logger.Debug("HEAD request failed for %s, assuming cache is valid: %v", urlStr, err)
		return true, nil
	}
	defer resp.Body.Close()

	// 304 Not Modified means cache is still fresh
	if resp.StatusCode == http.StatusNotModified {
		return true, nil
	}

	// Any other status means we should re-download
	return false, nil
}

// isTextContentType checks if a content type is text-based
func isTextContentType(contentType string) bool {
	// Remove parameters (e.g., "text/html; charset=utf-8" -> "text/html")
	contentType = strings.Split(contentType, ";")[0]
	contentType = strings.TrimSpace(strings.ToLower(contentType))

	// List of allowed text content types
	textTypes := []string{
		"text/",
		"application/json",
		"application/xml",
		"application/javascript",
		"application/x-yaml",
		"application/yaml",
	}

	for _, textType := range textTypes {
		if strings.HasPrefix(contentType, textType) {
			return true
		}
	}

	return false
}
