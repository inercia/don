package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/inercia/don/pkg/common"
	"github.com/inercia/don/pkg/utils"
	"gopkg.in/yaml.v3"
)

func TestConfigParsing(t *testing.T) {
	// Create a temporary config file
	tmpFile, err := os.CreateTemp("", "agent-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil && !os.IsNotExist(err) {
			t.Fatalf("Failed to remove temp file: %v", err)
		}
	}()

	configContent := `agent:
  models:
    - model: "test-model"
      class: "openai"
      name: "Test Agent"
      default: true
      api-key: "test-key"
      api-url: "https://api.test.com/v1"
      prompts:
        system:
        - "Test system prompt"

    - model: "test-model-2"
      class: "ollama"
      name: "Test Agent 2"
      default: false
      prompts:
        system:
        - "Test system prompt 2"
`

	_, err = tmpFile.WriteString(configContent)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Read and parse the config directly
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// Verify the config was loaded correctly
	if len(config.Agent.Models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(config.Agent.Models))
	}

	// Test GetDefaultModel
	defaultModel := config.GetDefaultModel()
	if defaultModel == nil {
		t.Fatal("Expected default model, got nil")
	}

	if defaultModel.Model != "test-model" {
		t.Errorf("Expected default model 'test-model', got '%s'", defaultModel.Model)
	}

	if !defaultModel.Default {
		t.Error("Expected default model to have Default=true")
	}

	if defaultModel.APIKey != "test-key" {
		t.Errorf("Expected API key 'test-key', got '%s'", defaultModel.APIKey)
	}

	// Test GetModelByName
	model := config.GetModelByName("test-model-2")
	if model == nil {
		t.Fatal("Expected to find model 'test-model-2', got nil")
	}

	if model.Model != "test-model-2" {
		t.Errorf("Expected model 'test-model-2', got '%s'", model.Model)
	}

	if model.Default {
		t.Error("Expected non-default model to have Default=false")
	}

	// Test GetModelByName with non-existent model
	nonExistentModel := config.GetModelByName("non-existent")
	if nonExistentModel != nil {
		t.Error("Expected nil for non-existent model")
	}
}

func TestEmptyConfig(t *testing.T) {
	config := Config{}

	// GetDefaultModel should return nil when no models
	defaultModel := config.GetDefaultModel()
	if defaultModel != nil {
		t.Error("Expected nil default model when no models configured")
	}

	// GetModelByName should return nil when no models
	model := config.GetModelByName("any-model")
	if model != nil {
		t.Error("Expected nil for any model when no models configured")
	}
}

func TestPromptsConfig(t *testing.T) {
	// Test empty prompts
	emptyPrompts := common.PromptsConfig{}

	if emptyPrompts.HasSystemPrompts() {
		t.Error("Expected false for HasSystemPrompts with empty config")
	}

	if emptyPrompts.HasUserPrompts() {
		t.Error("Expected false for HasUserPrompts with empty config")
	}

	if emptyPrompts.GetSystemPrompts() != "" {
		t.Error("Expected empty string for GetSystemPrompts with empty config")
	}

	if emptyPrompts.GetUserPrompts() != "" {
		t.Error("Expected empty string for GetUserPrompts with empty config")
	}

	// Test prompts with content
	prompts := common.PromptsConfig{
		System: []string{
			"You are a helpful assistant.",
			"Use available tools to help users.",
		},
		User: []string{
			"Help me with my task.",
			"Please be thorough.",
		},
	}

	if !prompts.HasSystemPrompts() {
		t.Error("Expected true for HasSystemPrompts with system prompts")
	}

	if !prompts.HasUserPrompts() {
		t.Error("Expected true for HasUserPrompts with user prompts")
	}

	expectedSystem := "You are a helpful assistant.\nUse available tools to help users."
	if prompts.GetSystemPrompts() != expectedSystem {
		t.Errorf("Expected system prompts '%s', got '%s'", expectedSystem, prompts.GetSystemPrompts())
	}

	expectedUser := "Help me with my task.\nPlease be thorough."
	if prompts.GetUserPrompts() != expectedUser {
		t.Errorf("Expected user prompts '%s', got '%s'", expectedUser, prompts.GetUserPrompts())
	}

	// Test single prompt
	singlePrompt := common.PromptsConfig{
		System: []string{"Single system prompt"},
	}

	if singlePrompt.GetSystemPrompts() != "Single system prompt" {
		t.Errorf("Expected 'Single system prompt', got '%s'", singlePrompt.GetSystemPrompts())
	}
}

func TestGetOrchestratorAndToolRunnerModels(t *testing.T) {
	// Test with role-based configuration
	config := Config{
		Agent: AgentConfigFile{
			Orchestrator: &ModelConfig{
				Model: "gpt-4o",
				Class: "openai",
				Name:  "orchestrator",
			},
			ToolRunner: &ModelConfig{
				Model: "gpt-4o-mini",
				Class: "openai",
				Name:  "tool-runner",
			},
		},
	}

	orchestrator := config.GetOrchestratorModel()
	if orchestrator == nil {
		t.Fatal("Expected orchestrator model, got nil")
	}
	if orchestrator.Model != "gpt-4o" {
		t.Errorf("Expected orchestrator model 'gpt-4o', got '%s'", orchestrator.Model)
	}

	toolRunner := config.GetToolRunnerModel()
	if toolRunner == nil {
		t.Fatal("Expected tool-runner model, got nil")
	}
	if toolRunner.Model != "gpt-4o-mini" {
		t.Errorf("Expected tool-runner model 'gpt-4o-mini', got '%s'", toolRunner.Model)
	}

	// Test with legacy flat model list
	legacyConfig := Config{
		Agent: AgentConfigFile{
			Models: []ModelConfig{
				{
					Model:   "gpt-4",
					Class:   "openai",
					Name:    "default",
					Default: true,
				},
			},
		},
	}

	legacyOrchestrator := legacyConfig.GetOrchestratorModel()
	if legacyOrchestrator == nil {
		t.Fatal("Expected orchestrator model from legacy config, got nil")
	}
	if legacyOrchestrator.Model != "gpt-4" {
		t.Errorf("Expected orchestrator model 'gpt-4', got '%s'", legacyOrchestrator.Model)
	}

	legacyToolRunner := legacyConfig.GetToolRunnerModel()
	if legacyToolRunner == nil {
		t.Fatal("Expected tool-runner model from legacy config, got nil")
	}
	// Tool runner should fall back to orchestrator
	if legacyToolRunner.Model != "gpt-4" {
		t.Errorf("Expected tool-runner to fall back to 'gpt-4', got '%s'", legacyToolRunner.Model)
	}

	// Test with empty config
	emptyConfig := Config{}
	emptyOrchestrator := emptyConfig.GetOrchestratorModel()
	if emptyOrchestrator != nil {
		t.Error("Expected nil orchestrator for empty config")
	}

	emptyToolRunner := emptyConfig.GetToolRunnerModel()
	if emptyToolRunner != nil {
		t.Error("Expected nil tool-runner for empty config")
	}
}

func TestGetConfigWithEnvVar(t *testing.T) {
	// Create a temporary config file
	tmpFile, err := os.CreateTemp("", "agent-config-env-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil && !os.IsNotExist(err) {
			t.Fatalf("Failed to remove temp file: %v", err)
		}
	}()

	configContent := `agent:
  models:
    - model: "test-env-model"
      class: "openai"
      name: "Test Env Agent"
      default: true
      api-key: "test-env-key"
`

	_, err = tmpFile.WriteString(configContent)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Save original env var value
	originalEnvValue := os.Getenv(utils.DonConfigEnv)
	defer func() {
		// Restore original env var value
		if originalEnvValue != "" {
			os.Setenv(utils.DonConfigEnv, originalEnvValue)
		} else {
			os.Unsetenv(utils.DonConfigEnv)
		}
	}()

	// Test 1: Set DON_CONFIG to custom path
	if err := os.Setenv(utils.DonConfigEnv, tmpFile.Name()); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}

	config, err := GetConfig()
	if err != nil {
		t.Fatalf("Failed to get config with env var: %v", err)
	}

	// Verify the config was loaded from the custom path
	if len(config.Agent.Models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(config.Agent.Models))
	}

	defaultModel := config.GetDefaultModel()
	if defaultModel == nil {
		t.Fatal("Expected default model, got nil")
	}

	if defaultModel.Model != "test-env-model" {
		t.Errorf("Expected model 'test-env-model', got '%s'", defaultModel.Model)
	}

	if defaultModel.APIKey != "test-env-key" {
		t.Errorf("Expected API key 'test-env-key', got '%s'", defaultModel.APIKey)
	}

	// Test 2: Unset env var and verify it uses default path
	if err := os.Unsetenv(utils.DonConfigEnv); err != nil {
		t.Fatalf("Failed to unset env var: %v", err)
	}

	// This should use the default path (~/.don/agent.yaml)
	// which may or may not exist, so we just verify it doesn't error
	_, err = GetConfig()
	if err != nil {
		t.Fatalf("Failed to get config without env var: %v", err)
	}
}

func TestGetConfigWithNonExistentEnvPath(t *testing.T) {
	// Save original env var value
	originalEnvValue := os.Getenv(utils.DonConfigEnv)
	defer func() {
		// Restore original env var value
		if originalEnvValue != "" {
			os.Setenv(utils.DonConfigEnv, originalEnvValue)
		} else {
			os.Unsetenv(utils.DonConfigEnv)
		}
	}()

	// Set env var to a non-existent path
	nonExistentPath := filepath.Join(os.TempDir(), "non-existent-config-12345.yaml")
	if err := os.Setenv(utils.DonConfigEnv, nonExistentPath); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}

	// GetConfig should return empty config when file doesn't exist
	config, err := GetConfig()
	if err != nil {
		t.Fatalf("Expected no error for non-existent config, got: %v", err)
	}

	// Should return empty config
	if len(config.Agent.Models) != 0 {
		t.Errorf("Expected empty config, got %d models", len(config.Agent.Models))
	}
}
