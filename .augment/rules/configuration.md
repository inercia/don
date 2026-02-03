# Configuration Standards for Don

## Agent Configuration

### Basic Agent Config Structure
```yaml
agent:
  models:
    - model: "gpt-4o"
      class: "openai"
      name: "GPT-4o Agent"
      default: true
      api-key: "${OPENAI_API_KEY}"
      api-url: "https://api.openai.com/v1"
      prompts:
        system:
          - "You are a helpful assistant."
```

### RAG Configuration
```yaml
agent:
  models:
    - model: "gpt-4o"
      class: "openai"
      default: true
      api-key: "${OPENAI_API_KEY}"

  rag:
    docs:
      description: "Project documentation"
      docs:
        - "https://example.com/docs"
        - "./docs"
      strategies:
        - type: "chunked-embeddings"
          model: "text-embedding-3-small"
          chunking:
            size: 1000
            overlap: 200
            respect_word_boundaries: true
          limit: 3
      results:
        limit: 5
        deduplicate: true
```

### Environment Variables in Config
- Use `${VAR_NAME}` syntax for environment variable substitution
- Common variables: `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, `GOOGLE_API_KEY`
- **NEVER** hardcode API keys in config files

### Agent Config Location
- Default: `~/.don/agent.yaml`
- Override with `DON_CONFIG` environment variable
- Use project-specific configs for different environments

## Tool Configuration

Don uses MCPShell's YAML format for tool configuration:

```yaml
mcp:
  description: "What this tool collection does"
  run:
    shell: bash
  tools:
    - name: "tool_name"
      description: "What the tool does"
      run:
        command: "echo {{ .param }}"
      params:
        param:
          type: string
          description: "Parameter description"
          required: true
```

### Tool Config Location
- Specified via `--tools` flag
- Can use MCPShell's tools directory (~/.mcpshell/tools)

## Supported Model Providers

### OpenAI
```yaml
- model: "gpt-4o"
  class: "openai"
  api-key: "${OPENAI_API_KEY}"
  api-url: "https://api.openai.com/v1"
```

### Anthropic
```yaml
- model: "claude-3-opus"
  class: "anthropic"
  api-key: "${ANTHROPIC_API_KEY}"
```

### Ollama (Local)
```yaml
- model: "llama3.2"
  class: "ollama"
  api-url: "http://localhost:11434"
```

### Amazon Bedrock
```yaml
- model: "anthropic.claude-3-sonnet"
  class: "bedrock"
  region: "us-east-1"
```

### Azure OpenAI
```yaml
- model: "gpt-4o"
  class: "azure"
  api-key: "${AZURE_OPENAI_KEY}"
  api-url: "https://your-resource.openai.azure.com"
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `DON_CONFIG` | Path to agent configuration file |
| `DON_DIR` | Path to Don's configuration directory |
| `DON_MODEL` | Default model to use |
| `OPENAI_API_KEY` | OpenAI API key |
| `ANTHROPIC_API_KEY` | Anthropic API key |
| `GOOGLE_API_KEY` | Google/Gemini API key |

## Validation

- Create config: `don config create`
- Show config info: `don info`
- Test with: `don --tools=tools.yaml "test prompt"`

