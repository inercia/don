package root

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for Don.

To load completions:

Bash:
  $ source <(don completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ don completion bash > /etc/bash_completion.d/don
  # macOS:
  $ don completion bash > $(brew --prefix)/etc/bash_completion.d/don

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ don completion zsh > "${fpath[1]}/_don"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ don completion fish | source

  # To load completions for each session, execute once:
  $ don completion fish > ~/.config/fish/completions/don.fish

PowerShell:
  PS> don completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> don completion powershell > don.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			_ = cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			_ = cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			_ = cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			_ = cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

// getMCPShellToolsDir returns the MCPShell tools directory
// Don uses MCPShell's tools configuration format
func getMCPShellToolsDir() string {
	if toolsDir := os.Getenv("MCPSHELL_TOOLS_DIR"); toolsDir != "" {
		return toolsDir
	}
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		return ""
	}
	return filepath.Join(home, ".mcpshell", "tools")
}

// listToolsFiles returns a list of available tools files for completion
func listToolsFiles() []string {
	var completions []string

	// Get tools from the MCPShell tools directory
	toolsDir := getMCPShellToolsDir()
	if toolsDir != "" {
		entries, err := os.ReadDir(toolsDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				name := entry.Name()
				if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
					// Add both with and without extension
					completions = append(completions, name)
					completions = append(completions, strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml"))
				}
			}
		}
	}

	// Get tools from current directory
	cwd, err := os.Getwd()
	if err == nil {
		entries, err := os.ReadDir(cwd)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				name := entry.Name()
				if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
					completions = append(completions, name)
				}
			}
		}
	}

	// Remove duplicates
	seen := make(map[string]bool)
	unique := make([]string, 0, len(completions))
	for _, c := range completions {
		if !seen[c] {
			seen[c] = true
			unique = append(unique, c)
		}
	}

	return unique
}

// toolsFileCompletion provides completion for the --tools flag
func toolsFileCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	completions := listToolsFiles()

	// Filter by prefix if user has typed something
	if toComplete != "" {
		filtered := make([]string, 0)
		for _, c := range completions {
			if strings.HasPrefix(c, toComplete) || strings.HasPrefix(filepath.Base(c), toComplete) {
				filtered = append(filtered, c)
			}
		}
		completions = filtered
	}

	// Also allow file completion for arbitrary paths
	return completions, cobra.ShellCompDirectiveDefault
}

func init() {
	rootCmd.AddCommand(completionCmd)

	// Register completion function for the --tools flag
	_ = rootCmd.RegisterFlagCompletionFunc("tools", toolsFileCompletion)
}
