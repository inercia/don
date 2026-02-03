// Package agent provides agent configuration and management functionality
package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/inercia/don/pkg/common"
	"github.com/inercia/don/pkg/rag"
)

// ProcessRAGSources processes RAG document sources (URLs, files, directories)
// Downloads remote URLs to local cache and scans local files/directories
// Returns a map of RAG source names to processed configurations with local paths
func ProcessRAGSources(ctx context.Context, ragSources map[string]RAGSourceConfig, logger *common.Logger) (map[string]RAGSourceConfig, error) {
	if len(ragSources) == 0 {
		return ragSources, nil
	}

	logger.Info("Processing RAG document sources...")

	// Create downloader for remote URLs
	downloaderConfig := rag.DownloaderConfig{
		Timeout: 30,               // 30 seconds
		MaxSize: 10 * 1024 * 1024, // 10MB
		Logger:  logger,
	}

	downloader, err := rag.NewDownloader(downloaderConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create document downloader: %w", err)
	}

	// Create file scanner for local files
	scanner := rag.NewScanner(logger)

	// Process each RAG source
	processedSources := make(map[string]RAGSourceConfig)
	for sourceName, sourceConfig := range ragSources {
		logger.Debug("Processing RAG source: %s", sourceName)

		// Process shared documents
		processedDocs, err := processDocuments(ctx, sourceConfig.Docs, downloader, scanner, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to process documents for RAG source '%s': %w", sourceName, err)
		}

		// Process strategy-specific documents
		processedStrategies := make([]RAGStrategyConfig, len(sourceConfig.Strategies))
		for i, strategy := range sourceConfig.Strategies {
			if len(strategy.Docs) > 0 {
				strategyDocs, err := processDocuments(ctx, strategy.Docs, downloader, scanner, logger)
				if err != nil {
					return nil, fmt.Errorf("failed to process documents for strategy '%s' in RAG source '%s': %w", strategy.Type, sourceName, err)
				}
				strategy.Docs = strategyDocs
			}
			processedStrategies[i] = strategy
		}

		// Create processed source config
		processedConfig := RAGSourceConfig{
			Description: sourceConfig.Description,
			Docs:        processedDocs,
			Strategies:  processedStrategies,
			Results:     sourceConfig.Results,
		}

		processedSources[sourceName] = processedConfig
		logger.Info("Processed RAG source '%s': %d shared documents, %d strategies", sourceName, len(processedDocs), len(processedStrategies))
	}

	return processedSources, nil
}

// processDocuments processes a list of document paths (URLs, files, directories)
// Downloads remote URLs and scans local files/directories
// Returns a list of local file paths
func processDocuments(ctx context.Context, docs []string, downloader rag.Downloader, scanner rag.Scanner, logger *common.Logger) ([]string, error) {
	var processedDocs []string

	for _, doc := range docs {
		// Check if it's a URL
		if strings.HasPrefix(doc, "http://") || strings.HasPrefix(doc, "https://") {
			// Download remote document
			logger.Debug("Downloading remote document: %s", doc)
			localPath, err := downloader.Download(ctx, doc)
			if err != nil {
				return nil, fmt.Errorf("failed to download document '%s': %w", doc, err)
			}
			processedDocs = append(processedDocs, localPath)
			logger.Debug("Downloaded to: %s", localPath)
		} else {
			// Process local file or directory
			absPath, err := filepath.Abs(doc)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve path '%s': %w", doc, err)
			}

			// Check if path exists
			info, err := os.Stat(absPath)
			if err != nil {
				return nil, fmt.Errorf("failed to access path '%s': %w", absPath, err)
			}

			if info.IsDir() {
				// Scan directory for text files
				logger.Debug("Scanning directory: %s", absPath)
				files, err := scanner.ScanDirectory(ctx, absPath, true) // recursive
				if err != nil {
					return nil, fmt.Errorf("failed to scan directory '%s': %w", absPath, err)
				}
				processedDocs = append(processedDocs, files...)
				logger.Debug("Found %d files in directory", len(files))
			} else {
				// Single file
				logger.Debug("Adding file: %s", absPath)
				processedDocs = append(processedDocs, absPath)
			}
		}
	}

	return processedDocs, nil
}
