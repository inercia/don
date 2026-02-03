// Package root provides the root command and CLI configuration for Don
package root

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/inercia/don/pkg/common"
)

var (
	// Version information set by build flags
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

// Global flags
var (
	logLevel   string
	logFile    string
	logToFile  bool
	toolsFiles []string
	verbose    bool
)

// Agent command flags
var (
	agentModel        string
	agentSystemPrompt string
	agentUserPrompt   string
	agentOpenAIApiKey string
	agentOpenAIApiURL string
	agentOnce         bool
	agentRAGSources   []string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "don",
	Short: "Don is an AI agent that connects LLMs to command-line tools",
	Long: `
Don is an AI agent that connects Large Language Models (LLMs) directly to
command-line tools, enabling autonomous task execution.

Don uses MCPShell tools configuration files to define available tools and
connects them with LLM providers like OpenAI, Anthropic, or Ollama.

Example:
  don --tools=tools.yaml "Help me debug this issue"
  don --tools=tools.yaml --model gpt-4o --once "What's the disk usage?"
`,
	Args: cobra.ArbitraryArgs,
}

// SetVersion sets version information from build flags
func SetVersion(v, c, d string) {
	version = v
	commit = c
	buildDate = d
}

// Execute runs the root command
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// initLogger initializes the logger based on command-line flags
func initLogger() (*common.Logger, error) {
	// Map verbose flag to log level
	level := common.LogLevelFromString(logLevel)
	if verbose {
		level = common.LogLevelDebug
	}

	// Create logger
	logger, err := common.NewLogger("", logFile, level, logToFile)
	if err != nil {
		return nil, err
	}

	// Set as global logger
	common.SetLogger(logger)

	return logger, nil
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error, none)")
	rootCmd.PersistentFlags().StringVar(&logFile, "log-file", "", "Log file path (default: stderr)")
	rootCmd.PersistentFlags().BoolVar(&logToFile, "log-to-file", false, "Write logs to file instead of stderr")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging (same as --log-level=debug)")
	rootCmd.PersistentFlags().StringSliceVarP(&toolsFiles, "tools", "t", []string{}, "Tool configuration file(s) (MCPShell format)")
}
