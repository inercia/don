# Don Documentation

Welcome to the Don documentation. Don is an AI agent that connects Large Language
Models (LLMs) directly to command-line tools, enabling autonomous task execution.

## Quick Links

- [Usage Guide](usage.md) - Getting started and running the agent
- [Configuration](configuration.md) - Agent and model configuration
- [RAG Guide](rag.md) - Using Retrieval-Augmented Generation
- [Architecture](architecture.md) - Technical architecture and design
- [Development](development.md) - Development guide and contributing
- [Release Process](release-process.md) - How to create releases

## Getting Started

1. **Install Don**

   ```bash
   go install github.com/inercia/don@latest
   ```

2. **Create configuration**

   ```bash
   don config create
   ```

   This creates `~/.don/agent.yaml` with sample model configurations.

3. **Set your API key**

   ```bash
   export OPENAI_API_KEY="sk-..."
   ```

4. **Run with tools**

   ```bash
   don --tools=tools.yaml "Help me with this task"
   ```

## Configuration

Don uses two types of configuration:

1. **Agent Configuration** (`~/.don/agent.yaml`) - Contains model definitions,
   API keys, and system prompts

2. **Tools Configuration** (any YAML file) - Contains tool definitions in
   MCPShell format that Don can execute

## Environment Variables

| Variable | Description |
|----------|-------------|
| `DON_CONFIG` | Path to agent configuration file (default: `~/.don/agent.yaml`) |
| `DON_DIR` | Path to Don's configuration directory (default: `~/.don`) |
| `DON_MODEL` | Default model to use |
| `OPENAI_API_KEY` | OpenAI API key |
| `ANTHROPIC_API_KEY` | Anthropic API key |
| `GOOGLE_API_KEY` | Google/Gemini API key |

## Supported Model Providers

- **OpenAI** - GPT-4, GPT-4o, GPT-3.5
- **Anthropic** - Claude models
- **Ollama** - Local LLMs (Llama, Qwen, Mistral, etc.)
- **Amazon Bedrock** - AWS-hosted models
- **Azure OpenAI** - Azure-hosted OpenAI models

## Requirements

Don requires MCPShell to be installed and available in your PATH for tool
execution. The tools configuration files are in MCPShell's YAML format.

Install MCPShell:

```bash
go install github.com/inercia/MCPShell@latest
```
