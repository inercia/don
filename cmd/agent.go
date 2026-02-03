package root

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/inercia/don/pkg/agent"
	"github.com/inercia/don/pkg/common"
)

// Cache the agent configuration to avoid duplicate resolution
var cachedAgentConfig agent.AgentConfig

// processArgsWithStdin processes positional arguments and replaces "-" with STDIN content
// Returns the processed prompt and a boolean indicating if STDIN was used
func processArgsWithStdin(args []string) (string, bool, error) {
	if len(args) == 0 {
		return "", false, nil
	}

	// Check if any argument is "-" (STDIN placeholder)
	hasStdin := false
	for _, arg := range args {
		if arg == "-" {
			hasStdin = true
			break
		}
	}

	// If no STDIN placeholder, just join the arguments
	if !hasStdin {
		return strings.Join(args, " "), false, nil
	}

	// Read STDIN content
	stdinContent, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", false, fmt.Errorf("failed to read STDIN: %w", err)
	}

	// Replace "-" with STDIN content in the arguments
	processedArgs := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "-" {
			processedArgs = append(processedArgs, string(stdinContent))
		} else {
			processedArgs = append(processedArgs, arg)
		}
	}

	return strings.Join(processedArgs, " "), true, nil
}

// buildAgentConfig creates an AgentConfig by merging command-line flags with configuration file
func buildAgentConfig() (agent.AgentConfig, error) {
	// Load configuration from file
	config, err := agent.GetConfig()
	if err != nil {
		return agent.AgentConfig{}, fmt.Errorf("failed to load config: %w", err)
	}

	// Start with default model from config file
	var modelConfig agent.ModelConfig
	if defaultModel := config.GetDefaultModel(); defaultModel != nil {
		modelConfig = *defaultModel
	}

	logger := common.GetLogger()

	// If --model flag not provided, check for environment variable
	if agentModel == "" {
		if envModel := os.Getenv("DON_MODEL"); envModel != "" {
			agentModel = envModel
			logger.Debug("Using model from DON_MODEL environment variable: %s", agentModel)
		}
	}

	// Override with command-line flags if provided
	if agentModel != "" {
		logger.Debug("Looking for model '%s' in agent config", agentModel)

		// Check if the specified model exists in config
		if configModel := config.GetModelByName(agentModel); configModel != nil {
			modelConfig = *configModel
			logger.Info("Found model '%s' in config: model=%s, class=%s, name=%s",
				agentModel, configModel.Model, configModel.Class, configModel.Name)
		} else {
			// Use command-line model name if not found in config
			logger.Info("Model '%s' not found in config, using as direct model name", agentModel)
			modelConfig.Model = agentModel
		}
	}

	// Merge system prompts from config file and command-line
	if agentSystemPrompt != "" {
		var allSystemPrompts []string
		if modelConfig.Prompts.HasSystemPrompts() {
			allSystemPrompts = append(allSystemPrompts, modelConfig.Prompts.System...)
		}
		allSystemPrompts = append(allSystemPrompts, agentSystemPrompt)
		modelConfig.Prompts.System = allSystemPrompts
	}

	// Override API key and URL if provided
	if agentOpenAIApiKey != "" {
		modelConfig.APIKey = agentOpenAIApiKey
	}
	if agentOpenAIApiURL != "" {
		modelConfig.APIURL = agentOpenAIApiURL
	}

	// Handle environment variable substitution for API key
	if strings.HasPrefix(modelConfig.APIKey, "${") && strings.HasSuffix(modelConfig.APIKey, "}") {
		envVar := strings.TrimSuffix(strings.TrimPrefix(modelConfig.APIKey, "${"), "}")
		modelConfig.APIKey = os.Getenv(envVar)
		logger.Debug("Substituted API key from environment variable: %s", envVar)
	}

	// Handle environment variable substitution for API URL
	if strings.HasPrefix(modelConfig.APIURL, "${") && strings.HasSuffix(modelConfig.APIURL, "}") {
		envVar := strings.TrimSuffix(strings.TrimPrefix(modelConfig.APIURL, "${"), "}")
		modelConfig.APIURL = os.Getenv(envVar)
		logger.Debug("Substituted API URL from environment variable: %s = %s", envVar, modelConfig.APIURL)
	}

	// Tools configuration is required
	if len(toolsFiles) == 0 {
		return agent.AgentConfig{}, fmt.Errorf("tools configuration file(s) are required (use --tools flag)")
	}

	// For don, we pass the tools files directly to the agent
	// The agent will spawn mcpshell to handle the tools
	localConfigPath := toolsFiles[0] // Use first tools file for now
	if len(toolsFiles) > 1 {
		logger.Warn("Multiple tools files not yet supported, using first: %s", localConfigPath)
	}

	// Build RAG configuration if sources are specified
	var ragConfig map[string]agent.RAGSourceConfig
	if len(agentRAGSources) > 0 {
		if len(config.Agent.RAG) == 0 {
			return agent.AgentConfig{}, fmt.Errorf("RAG sources specified but no RAG config in agent config file")
		}
		ragConfig = make(map[string]agent.RAGSourceConfig)
		for _, sourceName := range agentRAGSources {
			if ragSource, ok := config.Agent.RAG[sourceName]; ok {
				ragConfig[sourceName] = ragSource
				logger.Info("Enabled RAG source: %s", sourceName)
			} else {
				return agent.AgentConfig{}, fmt.Errorf("RAG source '%s' not found in agent config file", sourceName)
			}
		}
	}

	// Get the path to mcpshell binary - it must be available in PATH
	mcpshellBinary := "mcpshell"

	return agent.AgentConfig{
		ToolsFile:      localConfigPath,
		UserPrompt:     agentUserPrompt,
		Once:           agentOnce,
		Version:        version,
		MCPShellBinary: mcpshellBinary,
		ModelConfig:    modelConfig,
		RAGSources:     agentRAGSources,
		RAGConfig:      ragConfig,
	}, nil
}

// runAgent is the main agent execution logic
var runAgent = &cobra.Command{
	Use:   "don [flags] [prompt]",
	Short: "Run the Don AI agent",
	Long: `
Don is an AI agent that connects Large Language Models to command-line tools.

Configuration is loaded from ~/.don/agent.yaml and can be overridden with command-line flags.

Examples:
  don --tools=tools.yaml "Help me debug this issue"
  don -t tools --model gpt-4o "What's the disk usage?"
  don --tools=tools.yaml --once "Run the tests"

You can provide the initial prompt as positional arguments or use STDIN with '-':
  cat error.log | don --tools=tools.yaml "Analyze this error" -
`,
	Args: cobra.ArbitraryArgs,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// If --user-prompt is not provided but positional args exist, process them
		if agentUserPrompt == "" && len(args) > 0 {
			processedPrompt, usedStdin, err := processArgsWithStdin(args)
			if err != nil {
				return fmt.Errorf("failed to process arguments: %w", err)
			}
			agentUserPrompt = processedPrompt

			// If STDIN was used, automatically enable --once mode
			if usedStdin && !agentOnce {
				agentOnce = true
			}
		}

		// Initialize logger
		logger, err := initLogger()
		if err != nil {
			return err
		}

		// Build agent configuration (this will be cached for RunE)
		cachedAgentConfig, err = buildAgentConfig()
		if err != nil {
			return err
		}

		// Validate agent configuration
		agentInstance := agent.New(cachedAgentConfig, logger)
		if err := agentInstance.Validate(); err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger, err := initLogger()
		if err != nil {
			return err
		}

		agentConfig := cachedAgentConfig
		agentInstance := agent.New(agentConfig, logger)

		userInput := make(chan string)
		agentOutput := make(chan string)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-signalChan:
				logger.Info("Received interrupt signal, shutting down...")
				cancel()
			case <-ctx.Done():
			}
		}()

		// Start a goroutine to read user input only when not in --once mode
		if !agentConfig.Once {
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer close(userInput)

				scanner := bufio.NewScanner(os.Stdin)
				inputChan := make(chan string)

				go func() {
					for scanner.Scan() {
						inputChan <- scanner.Text()
					}
					close(inputChan)
				}()

				for {
					select {
					case <-ctx.Done():
						return
					case input, ok := <-inputChan:
						if !ok {
							return
						}
						select {
						case userInput <- input:
						case <-ctx.Done():
							return
						}
					}
				}
			}()
		}

		// Start the agent
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := agentInstance.Run(ctx, userInput, agentOutput); err != nil {
				if err != context.Canceled && err != context.DeadlineExceeded {
					logger.Error(color.HiRedString("Agent encountered an error: %v", err))
				}
				cancel()
			}
		}()

		// Print agent output
		for output := range agentOutput {
			fmt.Print(output)
		}

		// Wait for all goroutines with a timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			logger.Debug("All goroutines completed successfully")
		case <-time.After(5 * time.Second):
			logger.Debug("Cleanup timeout reached, forcing shutdown")
		}

		return nil
	},
}

func init() {
	// Set runAgent as the root command's run function
	rootCmd.PreRunE = runAgent.PreRunE
	rootCmd.RunE = runAgent.RunE

	// Add agent-specific flags
	rootCmd.PersistentFlags().StringVarP(&agentModel, "model", "m", "", "LLM model to use (can also set DON_MODEL env var)")
	rootCmd.PersistentFlags().StringVarP(&agentSystemPrompt, "system-prompt", "s", "", "System prompt for the LLM")
	rootCmd.PersistentFlags().StringVarP(&agentUserPrompt, "user-prompt", "u", "", "Initial user prompt for the LLM")
	rootCmd.PersistentFlags().StringVarP(&agentOpenAIApiKey, "api-key", "k", "", "API key (or set via environment variable)")
	rootCmd.PersistentFlags().StringVarP(&agentOpenAIApiURL, "api-url", "b", "", "Base URL for the API")
	rootCmd.PersistentFlags().BoolVarP(&agentOnce, "once", "o", false, "Exit after receiving a final response (one-shot mode)")
	rootCmd.PersistentFlags().StringSliceVar(&agentRAGSources, "rag", []string{}, "RAG source names to enable from config")
}
