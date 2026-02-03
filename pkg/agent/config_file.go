// Package agent provides agent configuration and management functionality
package agent

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/inercia/don/pkg/common"
	"github.com/inercia/don/pkg/utils"
)

//go:embed config_sample.yaml
var defaultConfigYAML string

// ModelConfig holds configuration for a single model
type ModelConfig struct {
	Model   string               `yaml:"model"`
	Class   string               `yaml:"class,omitempty"`   // Class of the model, e.g., "ollama", "openai", etc.
	Name    string               `yaml:"name,omitempty"`    // Name of the model, optional
	Default bool                 `yaml:"default,omitempty"` // Whether this is the default model
	APIKey  string               `yaml:"api-key,omitempty"` // API key, optional
	APIURL  string               `yaml:"api-url,omitempty"` // API URL, optional
	Prompts common.PromptsConfig `yaml:"prompts,omitempty"` // Prompts configuration, optional
}

// RAGChunkingConfig holds chunking configuration for RAG strategies
type RAGChunkingConfig struct {
	Size                  int  `yaml:"size,omitempty"`
	Overlap               int  `yaml:"overlap,omitempty"`
	RespectWordBoundaries bool `yaml:"respect_word_boundaries,omitempty"`
}

// RAGFusionConfig holds configuration for combining multi-strategy results
type RAGFusionConfig struct {
	Strategy string             `yaml:"strategy,omitempty"` // Fusion strategy: "rrf", "weighted", "max"
	K        int                `yaml:"k,omitempty"`        // RRF parameter k (default: 60)
	Weights  map[string]float64 `yaml:"weights,omitempty"`  // Strategy weights for weighted fusion
}

// RAGResultsConfig holds configuration for RAG result processing
type RAGResultsConfig struct {
	Limit             int              `yaml:"limit,omitempty"`               // Maximum number of results to return
	Fusion            *RAGFusionConfig `yaml:"fusion,omitempty"`              // How to combine results from multiple strategies
	Deduplicate       bool             `yaml:"deduplicate,omitempty"`         // Remove duplicate documents
	IncludeScore      bool             `yaml:"include_score,omitempty"`       // Include relevance scores
	ReturnFullContent bool             `yaml:"return_full_content,omitempty"` // Return full document content
}

// RAGStrategyConfig holds configuration for a single RAG retrieval strategy
type RAGStrategyConfig struct {
	Type     string            `yaml:"type"`               // Strategy type: "chunked-embeddings", "bm25"
	Docs     []string          `yaml:"docs,omitempty"`     // Strategy-specific documents
	Database string            `yaml:"database,omitempty"` // Database path for this strategy
	Chunking RAGChunkingConfig `yaml:"chunking,omitempty"` // Chunking configuration
	Limit    int               `yaml:"limit,omitempty"`    // Max results from this strategy

	// Strategy-specific parameters (e.g., model, threshold, vector_dimensions for chunked-embeddings)
	Model            string  `yaml:"model,omitempty"`
	Threshold        float64 `yaml:"threshold,omitempty"`
	VectorDimensions int     `yaml:"vector_dimensions,omitempty"`
	SimilarityMetric string  `yaml:"similarity_metric,omitempty"`
	K1               float64 `yaml:"k1,omitempty"` // BM25 parameter
	B                float64 `yaml:"b,omitempty"`  // BM25 parameter
}

// RAGSourceConfig holds configuration for a RAG knowledge source
type RAGSourceConfig struct {
	Description string              `yaml:"description"`
	Docs        []string            `yaml:"docs,omitempty"`       // Shared documents across all strategies
	Strategies  []RAGStrategyConfig `yaml:"strategies,omitempty"` // Array of strategy configurations
	Results     *RAGResultsConfig   `yaml:"results,omitempty"`
}

// AgentConfigFile holds the agent configuration from file
type AgentConfigFile struct {
	Models []ModelConfig `yaml:"models"` // Legacy: flat list of models

	// Role-based configuration for multi-agent system
	Orchestrator *ModelConfig `yaml:"orchestrator,omitempty"` // Root agent that plans and orchestrates
	ToolRunner   *ModelConfig `yaml:"tool-runner,omitempty"`  // Sub-agent that executes tools

	// RAG configuration
	RAG map[string]RAGSourceConfig `yaml:"rag,omitempty"` // Named RAG knowledge sources
}

// Config holds the complete agent configuration
type Config struct {
	Agent AgentConfigFile `yaml:"agent"`

	// Runtime fields (not from YAML)
	ToolsFile      string   // Path to tools configuration file
	RAGSources     []string // Names of RAG sources to use
	MCPShellBinary string   // Path to mcpshell binary (for spawning MCP server subprocess)
}

// GetConfig returns the agent configuration from the config file
// The config file location is determined by:
// 1. DON_CONFIG environment variable (if set)
// 2. Default: ~/.don/agent.yaml
func GetConfig() (*Config, error) {
	var configPath string

	// Check if DON_CONFIG environment variable is set
	if envConfigPath := os.Getenv(utils.DonConfigEnv); envConfigPath != "" {
		configPath = envConfigPath
	} else {
		// Use default location
		donHome, err := utils.GetDonHome()
		if err != nil {
			return nil, fmt.Errorf("failed to get Don home directory: %w", err)
		}
		configPath = filepath.Join(donHome, "agent.yaml")
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return empty config if file doesn't exist
		return &Config{}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	return &config, nil
}

// GetDefaultModel returns the model configuration that has default=true
// If no default is found, returns the first model in the list
// If no models are configured, returns nil
func (c *Config) GetDefaultModel() *ModelConfig {
	if len(c.Agent.Models) == 0 {
		return nil
	}

	// Look for the default model
	for i := range c.Agent.Models {
		if c.Agent.Models[i].Default {
			return &c.Agent.Models[i]
		}
	}

	// If no default found, return the first model
	return &c.Agent.Models[0]
}

// GetModelByName returns the model configuration with the specified name
func (c *Config) GetModelByName(name string) *ModelConfig {
	for i := range c.Agent.Models {
		// Check both Name and Model fields
		if c.Agent.Models[i].Name == name || c.Agent.Models[i].Model == name {
			return &c.Agent.Models[i]
		}
	}
	return nil
}

// CreateDefaultConfig creates a default agent configuration file if it doesn't exist
func CreateDefaultConfig() error {
	donHome, err := utils.GetDonHome()
	if err != nil {
		return fmt.Errorf("failed to get Don home directory: %w", err)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(donHome, 0o755); err != nil {
		return fmt.Errorf("failed to create Don directory: %w", err)
	}

	configPath := filepath.Join(donHome, "agent.yaml")

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		return nil // File already exists, don't overwrite
	}

	// Use the embedded default configuration
	if err := os.WriteFile(configPath, []byte(defaultConfigYAML), 0o644); err != nil {
		return fmt.Errorf("failed to write default config file: %w", err)
	}

	return nil
}

// CreateDefaultConfigForce creates a default agent configuration file, overwriting if it exists
func CreateDefaultConfigForce() error {
	donHome, err := utils.GetDonHome()
	if err != nil {
		return fmt.Errorf("failed to get Don home directory: %w", err)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(donHome, 0o755); err != nil {
		return fmt.Errorf("failed to create Don directory: %w", err)
	}

	configPath := filepath.Join(donHome, "agent.yaml")

	// Write the embedded default configuration
	if err := os.WriteFile(configPath, []byte(defaultConfigYAML), 0o644); err != nil {
		return fmt.Errorf("failed to write default config file: %w", err)
	}

	return nil
}

// GetDefaultConfig returns the default agent configuration parsed from the embedded config_sample.yaml
func GetDefaultConfig() (*Config, error) {
	var config Config
	if err := yaml.Unmarshal([]byte(defaultConfigYAML), &config); err != nil {
		return nil, fmt.Errorf("failed to parse default config: %w", err)
	}
	return &config, nil
}

// GetDefaultConfigYAML returns the embedded default configuration as a YAML string
func GetDefaultConfigYAML() string {
	return defaultConfigYAML
}

// GetOrchestratorModel returns the orchestrator model configuration
// Falls back to default model if orchestrator is not specified
func (c *Config) GetOrchestratorModel() *ModelConfig {
	if c.Agent.Orchestrator != nil {
		return c.Agent.Orchestrator
	}
	// Fall back to default model for backward compatibility
	return c.GetDefaultModel()
}

// GetToolRunnerModel returns the tool-runner model configuration
// Falls back to orchestrator model if tool-runner is not specified
func (c *Config) GetToolRunnerModel() *ModelConfig {
	if c.Agent.ToolRunner != nil {
		return c.Agent.ToolRunner
	}
	// Fall back to orchestrator model (which may fall back to default)
	return c.GetOrchestratorModel()
}
