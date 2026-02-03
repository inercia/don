// Package agent provides cagent runtime configuration and setup
package agent

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	cagentConfig "github.com/docker/cagent/pkg/config"
	"github.com/docker/cagent/pkg/runtime"
	"github.com/docker/cagent/pkg/session"
	"github.com/docker/cagent/pkg/teamloader"

	"github.com/inercia/don/pkg/common"
	"github.com/inercia/don/pkg/utils"
)

//go:embed prompts/orchestrator.md
var defaultOrchestratorPrompt string

// CagentRuntime wraps the cagent runtime and session
type CagentRuntime struct {
	runtime runtime.Runtime
	session *session.Session
	logger  *common.Logger
}

// CreateCagentRuntime creates and configures a cagent runtime using teamloader
// This enables full RAG support through cagent's built-in RAG system
// The MCP server is started as a subprocess, so the srv parameter is not needed
func CreateCagentRuntime(
	ctx context.Context,
	cfg *Config,
	userPrompt string,
	logger *common.Logger,
) (*CagentRuntime, error) {
	logger.Debug("Creating cagent runtime using teamloader")

	// Set up environment variables for API keys
	if err := setupEnvironment(cfg, logger); err != nil {
		return nil, fmt.Errorf("failed to setup environment: %w", err)
	}

	// Generate cagent-compatible YAML configuration
	yamlBytes, err := GenerateCagentYAML(cfg, cfg.ToolsFile, cfg.RAGSources, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to generate cagent config: %w", err)
	}

	logger.Debug("Generated cagent config YAML:\n%s", string(yamlBytes))

	// Write config to a temporary file
	// teamloader.Load requires a file path, not bytes
	donHome, err := utils.GetDonHome()
	if err != nil {
		return nil, fmt.Errorf("failed to get Don home: %w", err)
	}

	tempConfigPath := filepath.Join(donHome, ".cagent-runtime.yaml")
	if err := os.WriteFile(tempConfigPath, yamlBytes, 0600); err != nil {
		return nil, fmt.Errorf("failed to write temp config: %w", err)
	}
	defer os.Remove(tempConfigPath) // Clean up temp file

	logger.Debug("Wrote temp config to: %s", tempConfigPath)

	// Create runtime config
	runtimeConfig := cagentConfig.RuntimeConfig{}

	// Load team using teamloader - this handles RAG integration automatically
	logger.Debug("Loading team with teamloader (RAG will be initialized if configured)")
	agentTeam, err := teamloader.Load(ctx, cagentConfig.NewFileSource(tempConfigPath), &runtimeConfig)
	if err != nil {
		logger.Error("Failed to load team: %v", err)
		return nil, fmt.Errorf("failed to load team: %w", err)
	}

	logger.Debug("Team loaded successfully")

	// Create the runtime with session compaction enabled
	rt, err := runtime.New(
		agentTeam,
		runtime.WithSessionCompaction(true), // Auto-summarize when approaching context limit
	)
	if err != nil {
		logger.Error("Failed to create cagent runtime: %v", err)
		return nil, fmt.Errorf("failed to create cagent runtime: %w", err)
	}

	// Create the session with the user prompt
	// Enhance prompt to emphasize iterative workflow
	enhancedPrompt := userPrompt + `

Remember: This is a multi-step investigation. Keep calling tools iteratively until you have ALL the information needed to fully answer the question. Don't stop after just one tool call.`

	sess := session.New(session.WithUserMessage(enhancedPrompt))

	logger.Debug("Cagent runtime created successfully")

	return &CagentRuntime{
		runtime: rt,
		session: sess,
		logger:  logger,
	}, nil
}

// RunStream starts the streaming runtime and returns the event channel
func (cr *CagentRuntime) RunStream(ctx context.Context) <-chan runtime.Event {
	cr.logger.Debug("Starting cagent runtime stream")
	return cr.runtime.RunStream(ctx, cr.session)
}

// Runtime returns the underlying cagent runtime for advanced operations like Resume
func (cr *CagentRuntime) Runtime() runtime.Runtime {
	return cr.runtime
}

// ContinueConversation adds a new user message to the session and continues the conversation
func (cr *CagentRuntime) ContinueConversation(userMessage string) error {
	cr.logger.Debug("Adding user message to continue conversation")

	// Add the user message to the existing session
	msg := session.UserMessage(userMessage)
	cr.session.AddMessage(msg)

	cr.logger.Debug("User message added to session, ready for next stream")
	return nil
}

// setupEnvironment sets up environment variables for API keys from config
func setupEnvironment(cfg *Config, logger *common.Logger) error {
	// Set API keys for all models
	for _, model := range cfg.Agent.Models {
		if model.APIKey != "" {
			// Set the appropriate environment variable based on provider
			envVar := getAPIKeyEnvVar(model.Class)
			if err := os.Setenv(envVar, model.APIKey); err != nil {
				return fmt.Errorf("failed to set %s: %w", envVar, err)
			}
			logger.Debug("Set %s from model config", envVar)
		}
	}

	// Set orchestrator API key if specified
	if cfg.Agent.Orchestrator != nil && cfg.Agent.Orchestrator.APIKey != "" {
		envVar := getAPIKeyEnvVar(cfg.Agent.Orchestrator.Class)
		if err := os.Setenv(envVar, cfg.Agent.Orchestrator.APIKey); err != nil {
			return fmt.Errorf("failed to set %s: %w", envVar, err)
		}
		logger.Debug("Set %s from orchestrator config", envVar)
	}

	// Set tool-runner API key if specified
	if cfg.Agent.ToolRunner != nil && cfg.Agent.ToolRunner.APIKey != "" {
		envVar := getAPIKeyEnvVar(cfg.Agent.ToolRunner.Class)
		if err := os.Setenv(envVar, cfg.Agent.ToolRunner.APIKey); err != nil {
			return fmt.Errorf("failed to set %s: %w", envVar, err)
		}
		logger.Debug("Set %s from tool-runner config", envVar)
	}

	return nil
}

// getAPIKeyEnvVar returns the appropriate environment variable name for a provider
func getAPIKeyEnvVar(provider string) string {
	switch provider {
	case "anthropic":
		return "ANTHROPIC_API_KEY"
	case "google", "gemini":
		return "GOOGLE_API_KEY"
	case "ollama":
		return "OLLAMA_API_KEY" // Ollama doesn't need a key, but set it anyway
	default:
		return "OPENAI_API_KEY" // Default to OpenAI
	}
}
