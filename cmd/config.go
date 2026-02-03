package root

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/inercia/don/pkg/agent"
	"github.com/inercia/don/pkg/utils"
)

var (
	configShowJSON bool
)

// configCommand is the parent command for configuration subcommands
var configCommand = &cobra.Command{
	Use:   "config",
	Short: "Manage Don configuration",
	Long: `
The config command provides subcommands to manage Don configuration files.

Available subcommands:
- create: Create a default agent configuration file
- show: Display the current agent configuration
`,
}

// configCreateCommand creates a default configuration file
var configCreateCommand = &cobra.Command{
	Use:   "create",
	Short: "Create a default configuration file",
	Long: `
Creates a default agent configuration file at ~/.don/agent.yaml.

If the file already exists, it will be overwritten with the default configuration.

Example:
$ don config create
`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger, err := initLogger()
		if err != nil {
			return err
		}

		if err := agent.CreateDefaultConfigForce(); err != nil {
			logger.Error("Failed to create default config: %v", err)
			return fmt.Errorf("failed to create default config: %w", err)
		}

		donHome, err := utils.GetDonHome()
		if err != nil {
			return fmt.Errorf("failed to get Don home directory: %w", err)
		}

		configPath := filepath.Join(donHome, "agent.yaml")
		fmt.Printf("Default configuration created at: %s\n", configPath)
		fmt.Println("You can now edit this file to customize your agent settings.")

		return nil
	},
}

// ConfigShowOutput holds the JSON output structure for config show
type ConfigShowOutput struct {
	ConfigurationFile string                `json:"configuration_file"`
	Models            []ConfigShowModelInfo `json:"models"`
	DefaultModel      *ConfigShowModelInfo  `json:"default_model,omitempty"`
	Orchestrator      *ConfigShowModelInfo  `json:"orchestrator,omitempty"`
	ToolRunner        *ConfigShowModelInfo  `json:"tool_runner,omitempty"`
}

// ConfigShowModelInfo holds model info for JSON output
type ConfigShowModelInfo struct {
	Name          string   `json:"name"`
	Model         string   `json:"model"`
	Class         string   `json:"class"`
	Default       bool     `json:"default"`
	APIKey        string   `json:"api_key_masked,omitempty"`
	APIURL        string   `json:"api_url,omitempty"`
	SystemPrompts []string `json:"system_prompts,omitempty"`
}

// configShowCommand displays the current configuration
var configShowCommand = &cobra.Command{
	Use:   "show",
	Short: "Display the current configuration",
	Long: `
Displays the current agent configuration in a pretty-printed format.

Use --json flag to output in JSON format for easy parsing by other tools.

Examples:
$ don config show
$ don config show --json
`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger, err := initLogger()
		if err != nil {
			return err
		}

		donHome, err := utils.GetDonHome()
		if err != nil {
			return fmt.Errorf("failed to get Don home directory: %w", err)
		}
		configPath := filepath.Join(donHome, "agent.yaml")

		config, err := agent.GetConfig()
		if err != nil {
			logger.Error("Failed to load config: %v", err)
			return fmt.Errorf("failed to load config: %w", err)
		}

		if len(config.Agent.Models) == 0 {
			if configShowJSON {
				output := ConfigShowOutput{
					ConfigurationFile: configPath,
					Models:            []ConfigShowModelInfo{},
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(output)
			}

			fmt.Printf("Configuration file: %s\n", configPath)
			fmt.Println()
			fmt.Println("No agent configuration found.")
			fmt.Println("Run 'don config create' to create a default configuration.")
			return nil
		}

		if configShowJSON {
			return outputConfigShowJSON(configPath, config)
		}

		// Pretty print the configuration
		fmt.Printf("Configuration file: %s\n", configPath)
		fmt.Println()
		fmt.Println("Agent Configuration:")
		fmt.Println("===================")
		fmt.Println()

		for i, model := range config.Agent.Models {
			printModelInfo(i, model)
		}

		if defaultModel := config.GetDefaultModel(); defaultModel != nil {
			fmt.Printf("Default Model: %s (%s)\n", defaultModel.Name, defaultModel.Model)
		} else {
			fmt.Println("No default model configured.")
		}

		return nil
	},
}

func printModelInfo(index int, model agent.ModelConfig) {
	fmt.Printf("Model %d:\n", index+1)
	fmt.Printf("  Name: %s\n", model.Name)
	fmt.Printf("  Model: %s\n", model.Model)
	fmt.Printf("  Class: %s\n", model.Class)
	fmt.Printf("  Default: %t\n", model.Default)

	if model.APIKey != "" {
		if model.APIKey == "${OPENAI_API_KEY}" {
			fmt.Printf("  API Key: %s (from environment)\n", model.APIKey)
		} else {
			fmt.Printf("  API Key: %s\n", maskAPIKey(model.APIKey))
		}
	}

	if model.APIURL != "" {
		fmt.Printf("  API URL: %s\n", model.APIURL)
	}

	if model.Prompts.HasSystemPrompts() {
		systemPrompts := model.Prompts.GetSystemPrompts()
		fmt.Printf("  System Prompts: %s\n", truncateString(systemPrompts, 80))
	}

	fmt.Println()
}

func outputConfigShowJSON(configPath string, config *agent.Config) error {
	output := ConfigShowOutput{
		ConfigurationFile: configPath,
		Models:            make([]ConfigShowModelInfo, 0, len(config.Agent.Models)),
	}

	for _, model := range config.Agent.Models {
		modelInfo := ConfigShowModelInfo{
			Name:    model.Name,
			Model:   model.Model,
			Class:   model.Class,
			Default: model.Default,
			APIURL:  model.APIURL,
		}

		if model.APIKey != "" {
			modelInfo.APIKey = maskAPIKey(model.APIKey)
		}

		if model.Prompts.HasSystemPrompts() {
			modelInfo.SystemPrompts = model.Prompts.System
		}

		output.Models = append(output.Models, modelInfo)
	}

	if defaultModel := config.GetDefaultModel(); defaultModel != nil {
		modelInfo := ConfigShowModelInfo{
			Name:    defaultModel.Name,
			Model:   defaultModel.Model,
			Class:   defaultModel.Class,
			Default: defaultModel.Default,
			APIURL:  defaultModel.APIURL,
		}
		if defaultModel.APIKey != "" {
			modelInfo.APIKey = maskAPIKey(defaultModel.APIKey)
		}
		if defaultModel.Prompts.HasSystemPrompts() {
			modelInfo.SystemPrompts = defaultModel.Prompts.System
		}
		output.DefaultModel = &modelInfo
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func init() {
	rootCmd.AddCommand(configCommand)
	configCommand.AddCommand(configCreateCommand)
	configCommand.AddCommand(configShowCommand)

	configShowCommand.Flags().BoolVar(&configShowJSON, "json", false, "Output in JSON format")
}
