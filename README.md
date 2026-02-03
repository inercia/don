# Don

**Don** is an AI agent that connects Large Language Models (LLMs) directly to
command-line tools, enabling autonomous task execution without requiring a separate
MCP client.

## Features

- **Direct LLM connectivity**: Connect to OpenAI, Anthropic, Ollama, and other
  LLM providers
- **Multi-agent architecture**: Uses orchestrator and tool-runner agents for
  complex task execution
- **RAG support**: Retrieval-Augmented Generation for document-based Q&A
- **Flexible configuration**: YAML-based configuration with environment variable
  substitution
- **Multiple retrieval strategies**: Chunked embeddings, BM25, and hybrid search

## Quick Start

1. Create an agent configuration file at `~/.don/agent.yaml`:

   ```yaml
   agent:
     models:
       - model: "gpt-4o"
         class: "openai"
         name: "gpt-4o"
         default: true
         api-key: "${OPENAI_API_KEY}"
         api-url: "https://api.openai.com/v1"
   ```

2. Set your API key:

   ```bash
   export OPENAI_API_KEY="sk-..."
   ```

3. Run the agent with a tools configuration:

   ```bash
   don --tools=examples/tools.yaml "Help me debug this issue"
   ```

## Installation

```bash
go install github.com/inercia/don@latest
```

Or build from source:

```bash
git clone https://github.com/inercia/don
cd don
make build
```

## Usage

### Basic Usage

```bash
# Run with a specific model
don --tools=tools.yaml --model gpt-4o "Your question here"

# One-shot mode (exit after response)
don --tools=tools.yaml --once "What's the disk usage?"

# Interactive mode
don --tools=tools.yaml
```

### With RAG

```bash
# Enable RAG sources from config
don --tools=tools.yaml --rag=docs "What does the documentation say about X?"
```

### Subcommands

```bash
# Show agent configuration
don info

# Create default configuration
don config create

# Show current configuration
don config show
```

## Configuration

See the [Configuration Guide](docs/configuration.md) for detailed configuration
options.

## Documentation

- [Usage Guide](docs/usage.md) - Getting started and basic usage
- [Configuration](docs/configuration.md) - Agent and model configuration
- [RAG Guide](docs/rag.md) - Using Retrieval-Augmented Generation
- [Architecture](docs/architecture.md) - Technical architecture and design

## License

This project is licensed under the MIT License - see the LICENSE file for details.

