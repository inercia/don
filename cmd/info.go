package root

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"

	"github.com/inercia/don/pkg/agent"
	"github.com/inercia/don/pkg/common"
	"github.com/inercia/don/pkg/utils"
)

var (
	infoJSON           bool
	infoIncludePrompts bool
	infoCheck          bool
)

// infoCommand displays information about the agent configuration
var infoCommand = &cobra.Command{
	Use:   "info",
	Short: "Display agent configuration information",
	Long: `
Display information about the agent configuration including:
- LLM model details
- API configuration
- System prompts (with --include-prompts)
- LLM connectivity status (with --check)

Examples:
$ don info
$ don info --json
$ don info --include-prompts
$ don info --check
$ don info --model gpt-4o --json
`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger, err := initLogger()
		if err != nil {
			return err
		}

		agentConfig, err := buildAgentConfigForInfo()
		if err != nil {
			return fmt.Errorf("failed to build agent config: %w", err)
		}

		orchestratorConfig := agentConfig.ModelConfig
		toolRunnerConfig := agentConfig.ModelConfig

		var checkResult *CheckResult
		if infoCheck {
			checkResult = checkLLMConnectivity(orchestratorConfig, logger)
		}

		if infoJSON {
			err := outputInfoJSON(agentConfig, orchestratorConfig, toolRunnerConfig, checkResult)
			if err != nil {
				return err
			}
			if checkResult != nil && !checkResult.Success {
				return fmt.Errorf("LLM connectivity check failed: %s", checkResult.Error)
			}
			return nil
		}

		return outputInfoHumanReadable(agentConfig, orchestratorConfig, toolRunnerConfig, checkResult)
	},
}

// CheckResult holds the result of an LLM connectivity check
type CheckResult struct {
	Success      bool    `json:"success"`
	ResponseTime float64 `json:"response_time_ms"`
	Error        string  `json:"error,omitempty"`
	Model        string  `json:"model"`
}

// InfoOutput holds the complete info output structure for JSON
type InfoOutput struct {
	ConfigFile   string       `json:"config_file,omitempty"`
	ToolsFile    string       `json:"tools_file,omitempty"`
	Once         bool         `json:"once_mode"`
	Orchestrator ModelInfo    `json:"orchestrator"`
	ToolRunner   ModelInfo    `json:"tool_runner"`
	Check        *CheckResult `json:"check,omitempty"`
	Prompts      *PromptsInfo `json:"prompts,omitempty"`
}

// ModelInfo holds model configuration details for JSON output
type ModelInfo struct {
	Model  string `json:"model"`
	Class  string `json:"class"`
	Name   string `json:"name,omitempty"`
	APIURL string `json:"api_url,omitempty"`
	APIKey string `json:"api_key_masked,omitempty"`
}

// PromptsInfo holds prompt information for JSON output
type PromptsInfo struct {
	System []string `json:"system,omitempty"`
	User   string   `json:"user,omitempty"`
}

// buildAgentConfigForInfo creates an AgentConfig for the info command
func buildAgentConfigForInfo() (agent.AgentConfig, error) {
	config, err := agent.GetConfig()
	if err != nil {
		return agent.AgentConfig{}, fmt.Errorf("failed to load config: %w", err)
	}

	var modelConfig agent.ModelConfig
	if defaultModel := config.GetDefaultModel(); defaultModel != nil {
		modelConfig = *defaultModel
	}

	logger := common.GetLogger()

	if agentModel == "" {
		if envModel := os.Getenv("DON_MODEL"); envModel != "" {
			agentModel = envModel
			logger.Debug("Using model from DON_MODEL environment variable: %s", agentModel)
		}
	}

	if agentModel != "" {
		if configModel := config.GetModelByName(agentModel); configModel != nil {
			modelConfig = *configModel
		} else {
			modelConfig.Model = agentModel
		}
	}

	if agentSystemPrompt != "" {
		var allSystemPrompts []string
		if modelConfig.Prompts.HasSystemPrompts() {
			allSystemPrompts = append(allSystemPrompts, modelConfig.Prompts.System...)
		}
		allSystemPrompts = append(allSystemPrompts, agentSystemPrompt)
		modelConfig.Prompts.System = allSystemPrompts
	}

	if agentOpenAIApiKey != "" {
		modelConfig.APIKey = agentOpenAIApiKey
	}
	if agentOpenAIApiURL != "" {
		modelConfig.APIURL = agentOpenAIApiURL
	}

	if strings.HasPrefix(modelConfig.APIKey, "${") && strings.HasSuffix(modelConfig.APIKey, "}") {
		envVar := strings.TrimSuffix(strings.TrimPrefix(modelConfig.APIKey, "${"), "}")
		modelConfig.APIKey = os.Getenv(envVar)
	}

	if strings.HasPrefix(modelConfig.APIURL, "${") && strings.HasSuffix(modelConfig.APIURL, "}") {
		envVar := strings.TrimSuffix(strings.TrimPrefix(modelConfig.APIURL, "${"), "}")
		modelConfig.APIURL = os.Getenv(envVar)
	}

	toolsFile := ""
	if len(toolsFiles) > 0 {
		toolsFile = toolsFiles[0]
	}

	return agent.AgentConfig{
		ToolsFile:   toolsFile,
		UserPrompt:  agentUserPrompt,
		Once:        agentOnce,
		Version:     version,
		ModelConfig: modelConfig,
	}, nil
}

func checkLLMConnectivity(modelConfig agent.ModelConfig, logger *common.Logger) *CheckResult {
	result := &CheckResult{
		Model: modelConfig.Model,
	}

	logger.Info("Testing LLM connectivity for model: %s", modelConfig.Model)

	client, err := agent.InitializeModelClient(modelConfig, logger)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Failed to initialize client: %v", err)
		return result
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	startTime := time.Now()

	req := openai.ChatCompletionRequest{
		Model: modelConfig.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Respond with just the word 'OK'",
			},
		},
		MaxTokens: 10,
	}

	_, err = client.CreateChatCompletion(ctx, req)
	elapsed := time.Since(startTime)

	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("LLM request failed: %v", err)
		logger.Error("LLM connectivity check failed: %v", err)
		return result
	}

	result.Success = true
	result.ResponseTime = float64(elapsed.Milliseconds())
	logger.Info("LLM connectivity check successful (%.0fms)", result.ResponseTime)

	return result
}

func outputInfoJSON(agentConfig agent.AgentConfig, orchestrator, toolRunner agent.ModelConfig, check *CheckResult) error {
	var configFile string
	if donHome, err := utils.GetDonHome(); err == nil {
		configFile = filepath.Join(donHome, "agent.yaml")
	}

	output := InfoOutput{
		ConfigFile: configFile,
		ToolsFile:  agentConfig.ToolsFile,
		Once:       agentConfig.Once,
		Orchestrator: ModelInfo{
			Model:  orchestrator.Model,
			Class:  orchestrator.Class,
			Name:   orchestrator.Name,
			APIURL: orchestrator.APIURL,
			APIKey: maskAPIKey(orchestrator.APIKey),
		},
		ToolRunner: ModelInfo{
			Model:  toolRunner.Model,
			Class:  toolRunner.Class,
			Name:   toolRunner.Name,
			APIURL: toolRunner.APIURL,
			APIKey: maskAPIKey(toolRunner.APIKey),
		},
		Check: check,
	}

	if infoIncludePrompts {
		output.Prompts = &PromptsInfo{
			System: orchestrator.Prompts.System,
			User:   agentConfig.UserPrompt,
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputInfoHumanReadable(agentConfig agent.AgentConfig, orchestrator, toolRunner agent.ModelConfig, check *CheckResult) error {
	fmt.Println(color.HiCyanString("Don Agent Configuration"))
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	donHome, err := utils.GetDonHome()
	if err == nil {
		agentConfigPath := filepath.Join(donHome, "agent.yaml")
		fmt.Printf("Config File:   %s\n", agentConfigPath)
	}

	if agentConfig.ToolsFile != "" {
		fmt.Printf("Tools File:    %s\n", agentConfig.ToolsFile)
	}
	fmt.Printf("Once Mode:     %t\n", agentConfig.Once)
	fmt.Println()

	fmt.Println(color.HiYellowString("Orchestrator Model:"))
	fmt.Printf("  Model:       %s\n", orchestrator.Model)
	if orchestrator.Name != "" {
		fmt.Printf("  Name:        %s\n", orchestrator.Name)
	}
	fmt.Printf("  Class:       %s\n", orchestrator.Class)
	if orchestrator.APIURL != "" {
		fmt.Printf("  API URL:     %s\n", orchestrator.APIURL)
	}
	if orchestrator.APIKey != "" {
		fmt.Printf("  API Key:     %s\n", maskAPIKey(orchestrator.APIKey))
	}
	fmt.Println()

	if infoIncludePrompts {
		fmt.Println(color.HiYellowString("Prompts:"))
		if orchestrator.Prompts.HasSystemPrompts() {
			fmt.Println(color.CyanString("  System Prompts:"))
			for i, prompt := range orchestrator.Prompts.System {
				fmt.Printf("    %d. %s\n", i+1, truncateString(prompt, 120))
			}
		} else {
			fmt.Println("  System Prompts: (none)")
		}
		if agentConfig.UserPrompt != "" {
			fmt.Printf("  User Prompt:   %s\n", truncateString(agentConfig.UserPrompt, 120))
		}
		fmt.Println()
	}

	if check != nil {
		fmt.Println(color.HiYellowString("LLM Connectivity Check:"))
		if check.Success {
			fmt.Printf("  Status:      %s\n", color.HiGreenString("✓ Connected"))
			fmt.Printf("  Response:    %.0fms\n", check.ResponseTime)
		} else {
			fmt.Printf("  Status:      %s\n", color.HiRedString("✗ Failed"))
			fmt.Printf("  Error:       %s\n", check.Error)
			return fmt.Errorf("LLM connectivity check failed: %s", check.Error)
		}
		fmt.Println()
	}

	return nil
}

func init() {
	rootCmd.AddCommand(infoCommand)

	infoCommand.Flags().BoolVar(&infoJSON, "json", false, "Output in JSON format")
	infoCommand.Flags().BoolVar(&infoIncludePrompts, "include-prompts", false, "Include full prompts in the output")
	infoCommand.Flags().BoolVar(&infoCheck, "check", false, "Check LLM connectivity")
}
