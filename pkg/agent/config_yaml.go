// Package agent provides agent configuration and management functionality
package agent

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/inercia/don/pkg/common"
)

// GenerateCagentYAML generates a cagent-compatible YAML configuration
// from our MCPShell configuration
func GenerateCagentYAML(
	cfg *Config,
	toolsFile string,
	ragSources []string,
	logger *common.Logger,
) ([]byte, error) {
	logger.Debug("Generating cagent YAML configuration")

	// Build the config structure as a map for easy YAML generation
	cagentCfg := make(map[string]interface{})
	cagentCfg["version"] = "v2"

	// Convert models
	models := make(map[string]interface{})
	if err := addModels(cfg, models, logger); err != nil {
		return nil, fmt.Errorf("failed to add models: %w", err)
	}
	if len(models) > 0 {
		cagentCfg["models"] = models
	}

	// Convert RAG sources
	if len(cfg.Agent.RAG) > 0 {
		rag := make(map[string]interface{})
		if err := addRAGSources(cfg, rag, logger); err != nil {
			return nil, fmt.Errorf("failed to add RAG sources: %w", err)
		}
		cagentCfg["rag"] = rag
	}

	// Create root agent
	agents := make(map[string]interface{})
	if err := addRootAgent(cfg, agents, toolsFile, ragSources, logger); err != nil {
		return nil, fmt.Errorf("failed to add root agent: %w", err)
	}
	cagentCfg["agents"] = agents

	// Marshal to YAML
	yamlBytes, err := yaml.Marshal(cagentCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	logger.Debug("Generated cagent YAML configuration (%d bytes)", len(yamlBytes))
	return yamlBytes, nil
}

// addModels adds model configurations to the cagent config
func addModels(cfg *Config, models map[string]interface{}, logger *common.Logger) error {
	// Add all models from the flat list
	for _, model := range cfg.Agent.Models {
		name := model.Name
		if name == "" {
			name = model.Model
		}

		models[name] = map[string]interface{}{
			"provider":  model.Class,
			"model":     model.Model,
			"base_url":  model.APIURL,
			"token_key": model.APIKey,
		}

		logger.Debug("Added model: %s (provider: %s, model: %s)",
			name, model.Class, model.Model)
	}

	// Add orchestrator if specified
	if cfg.Agent.Orchestrator != nil {
		name := "orchestrator"
		if cfg.Agent.Orchestrator.Name != "" {
			name = cfg.Agent.Orchestrator.Name
		}

		models[name] = map[string]interface{}{
			"provider":  cfg.Agent.Orchestrator.Class,
			"model":     cfg.Agent.Orchestrator.Model,
			"base_url":  cfg.Agent.Orchestrator.APIURL,
			"token_key": cfg.Agent.Orchestrator.APIKey,
		}

		logger.Debug("Added orchestrator model: %s", name)
	}

	// Add tool-runner if specified
	if cfg.Agent.ToolRunner != nil {
		name := "tool-runner"
		if cfg.Agent.ToolRunner.Name != "" {
			name = cfg.Agent.ToolRunner.Name
		}

		models[name] = map[string]interface{}{
			"provider":  cfg.Agent.ToolRunner.Class,
			"model":     cfg.Agent.ToolRunner.Model,
			"base_url":  cfg.Agent.ToolRunner.APIURL,
			"token_key": cfg.Agent.ToolRunner.APIKey,
		}

		logger.Debug("Added tool-runner model: %s", name)
	}

	if len(models) == 0 {
		return fmt.Errorf("no models configured")
	}

	return nil
}

// addRAGSources adds RAG configurations to the cagent config
func addRAGSources(cfg *Config, rag map[string]interface{}, logger *common.Logger) error {
	for name, ragSrc := range cfg.Agent.RAG {
		ragCfg := make(map[string]interface{})

		// Add tool configuration
		tool := make(map[string]interface{})
		tool["name"] = name
		tool["description"] = ragSrc.Description
		tool["instruction"] = fmt.Sprintf("Search %s for relevant information", ragSrc.Description)
		ragCfg["tool"] = tool

		// Add docs
		if len(ragSrc.Docs) > 0 {
			ragCfg["docs"] = ragSrc.Docs
		}

		// Add strategies
		if len(ragSrc.Strategies) > 0 {
			strategies := make([]interface{}, len(ragSrc.Strategies))
			for i, s := range ragSrc.Strategies {
				strategies[i] = convertStrategy(s)
			}
			ragCfg["strategies"] = strategies
		}

		// Add results configuration
		if ragSrc.Results != nil {
			results := convertResults(ragSrc.Results)
			ragCfg["results"] = results
		}

		rag[name] = ragCfg

		logger.Debug("Added RAG source: %s (docs: %d, strategies: %d)",
			name, len(ragSrc.Docs), len(ragSrc.Strategies))
	}

	return nil
}

// convertStrategy converts a RAG strategy to a map for YAML generation
func convertStrategy(s RAGStrategyConfig) map[string]interface{} {
	strategy := make(map[string]interface{})
	strategy["type"] = s.Type

	if len(s.Docs) > 0 {
		strategy["docs"] = s.Docs
	}

	if s.Limit > 0 {
		strategy["limit"] = s.Limit
	}

	// Add chunking configuration
	if s.Chunking.Size > 0 || s.Chunking.Overlap > 0 {
		chunking := make(map[string]interface{})
		if s.Chunking.Size > 0 {
			chunking["size"] = s.Chunking.Size
		}
		if s.Chunking.Overlap > 0 {
			chunking["overlap"] = s.Chunking.Overlap
		}
		chunking["respect_word_boundaries"] = s.Chunking.RespectWordBoundaries
		strategy["chunking"] = chunking
	}

	// Add database configuration
	if s.Database != "" {
		strategy["database"] = s.Database
	}

	// Add strategy-specific parameters (flattened into the strategy map)
	if s.Model != "" {
		strategy["model"] = s.Model
	}
	if s.Threshold != 0 {
		strategy["threshold"] = s.Threshold
	}
	if s.VectorDimensions != 0 {
		strategy["vector_dimensions"] = s.VectorDimensions
	}
	if s.SimilarityMetric != "" {
		strategy["similarity_metric"] = s.SimilarityMetric
	}
	if s.K1 != 0 {
		strategy["k1"] = s.K1
	}
	if s.B != 0 {
		strategy["b"] = s.B
	}

	return strategy
}

// convertResults converts RAG results configuration to a map for YAML generation
func convertResults(results *RAGResultsConfig) map[string]interface{} {
	r := make(map[string]interface{})

	if results.Limit > 0 {
		r["limit"] = results.Limit
	}
	r["deduplicate"] = results.Deduplicate
	r["include_score"] = results.IncludeScore
	r["return_full_content"] = results.ReturnFullContent

	// Add fusion configuration
	if results.Fusion != nil {
		fusion := make(map[string]interface{})
		if results.Fusion.Strategy != "" {
			fusion["strategy"] = results.Fusion.Strategy
		}
		if results.Fusion.K > 0 {
			fusion["k"] = results.Fusion.K
		}
		if len(results.Fusion.Weights) > 0 {
			fusion["weights"] = results.Fusion.Weights
		}
		r["fusion"] = fusion
	}

	return r
}

// addRootAgent adds the root agent configuration to the cagent config
func addRootAgent(
	cfg *Config,
	agents map[string]interface{},
	toolsFile string,
	ragSources []string,
	logger *common.Logger,
) error {
	// Get the default model name
	defaultModel := cfg.GetDefaultModel()
	if defaultModel == nil {
		return fmt.Errorf("no default model found")
	}

	modelName := defaultModel.Name
	if modelName == "" {
		modelName = defaultModel.Model
	}

	// Get system prompt
	systemPrompt := getSystemPrompt(cfg)

	// Create root agent
	rootAgent := make(map[string]interface{})
	rootAgent["model"] = modelName
	rootAgent["instruction"] = systemPrompt

	// Add toolsets - MCP server as a toolset
	// Use command to start mcpshell as an MCP server subprocess
	if toolsFile != "" {
		// Use the configured binary path, or default to "mcpshell" if not set
		mcpBinary := "mcpshell"
		if cfg.MCPShellBinary != "" {
			mcpBinary = cfg.MCPShellBinary
		}

		// cagent expects separate command and args fields for MCP toolsets
		// This matches the format used by mcp-go's NewStdioMCPClient
		toolsets := []interface{}{
			map[string]interface{}{
				"type":    "mcp",
				"command": mcpBinary,
				"args":    []string{"mcp", "--tools", toolsFile},
			},
		}
		rootAgent["toolsets"] = toolsets
		logger.Debug("Added MCP toolset with command: %s, args: [mcp --tools %s]", mcpBinary, toolsFile)
	}

	// Add RAG sources
	if len(ragSources) > 0 {
		rootAgent["rag"] = ragSources
		logger.Debug("Added RAG sources: %s", strings.Join(ragSources, ", "))
	}

	agents["root"] = rootAgent

	logger.Debug("Created root agent: model=%s, RAG sources=%d",
		modelName, len(ragSources))

	return nil
}

// getSystemPrompt returns the system prompt for the agent
func getSystemPrompt(cfg *Config) string {
	// Check if orchestrator has custom prompts
	if cfg.Agent.Orchestrator != nil && len(cfg.Agent.Orchestrator.Prompts.System) > 0 {
		return cfg.Agent.Orchestrator.Prompts.System[0]
	}

	// Check default model for prompts
	defaultModel := cfg.GetDefaultModel()
	if defaultModel != nil && len(defaultModel.Prompts.System) > 0 {
		return defaultModel.Prompts.System[0]
	}

	// Default system prompt
	return "You are a helpful AI assistant with access to command-line tools via MCP (Model Context Protocol). " +
		"Use the available tools to help users accomplish their tasks safely and effectively."
}
