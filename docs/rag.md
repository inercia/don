# Using Agent Mode with RAG

This guide shows you how to use Don's agent mode with RAG (Retrieval-Augmented
Generation) to create agents that can answer questions based on your documentation.

## What is RAG?

RAG allows your agent to retrieve relevant information from documents (URLs, local
files, or directories) to answer questions accurately. Instead of relying solely on the
LLM's training data, the agent can search through your specific documentation and use
that information to provide answers.

## Prerequisites

- Don installed and configured
- OpenAI API key (for embeddings and LLM)
- Documents you want the agent to use (URLs, local files, or directories)

## Golden Path: Quick Start

This is the fastest way to get started with RAG.

### Step 1: Create Agent Configuration

Create or edit your agent configuration file at `~/.don/agent.yaml`:

```yaml
agent:
  # Configure your LLM model
  models:
    - model: "gpt-4o"
      class: "openai"
      name: "openai"
      default: true
      api-key: "${OPENAI_API_KEY}"
      api-url: "https://api.openai.com/v1"

  # Define RAG sources
  rag:
    docs:
      description: "Project documentation"

      # Add your documents (URLs or local paths)
      docs:
        - "https://raw.githubusercontent.com/inercia/Don/main/README.md"
        - "./docs"

      # Configure retrieval strategy
      strategies:
        - type: "chunked-embeddings"
          model: "text-embedding-3-small"
          chunking:
            size: 1000
            overlap: 200
          limit: 5

      # Configure results
      results:
        limit: 10
        deduplicate: true
```

### Step 2: Set Your API Key

```bash
export OPENAI_API_KEY="sk-your-api-key-here"
```

### Step 3: Run the Agent with RAG

```bash
don agent --tools=tools.yaml --rag=docs "What is Don?"
```

That's it! The agent will:

1. Download the README from GitHub (cached locally)
2. Scan your local `./docs` directory
3. Index the documents with embeddings
4. Answer your question using information from the documents

## How It Works

When you run the agent with RAG:

1. **Document Processing**: Don downloads URLs and scans local paths
   - URLs are cached in `~/.cache/don/rag/` (or platform-specific location)
   - Local files are scanned for text content
   - All documents are converted to local paths

2. **Indexing** (when full integration is complete):
   - Documents are split into chunks
   - Each chunk is embedded using the specified model
   - Embeddings are stored for fast retrieval

3. **Query Processing**:
   - Your question is embedded
   - Similar chunks are retrieved from the index
   - The LLM uses the retrieved information to answer

4. **Response**:
   - The agent provides an answer based on your documents
   - More accurate and up-to-date than relying on training data alone

## Configuration Options

### Document Sources

You can specify multiple types of document sources:

```yaml
docs:
  # URLs (downloaded and cached)
  - "https://example.com/docs/getting-started.md"
  - "https://example.com/api/reference.html"

  # Local files
  - "./README.md"
  - "./CHANGELOG.md"

  # Local directories (scanned recursively)
  - "./docs"
  - "./examples"
```

### Retrieval Strategies

#### Chunked Embeddings (Recommended)

Vector similarity search using embeddings:

```yaml
strategies:
  - type: "chunked-embeddings"
    model: "text-embedding-3-small" # OpenAI embedding model
    chunking:
      size: 1000 # Chunk size in characters
      overlap: 200 # Overlap between chunks
      respect_word_boundaries: true # Don't split words
    limit: 5 # Max chunks per query
```

#### BM25 (Keyword Search)

Traditional keyword-based search:

```yaml
strategies:
  - type: "bm25"
    limit: 5
```

#### Multiple Strategies with Fusion

Combine multiple strategies for better results:

```yaml
strategies:
  - type: "chunked-embeddings"
    model: "text-embedding-3-small"
    limit: 3

  - type: "bm25"
    limit: 3

results:
  limit: 5
  fusion:
    strategy: "rrf" # Reciprocal Rank Fusion
    k: 60
  deduplicate: true
```

### Result Configuration

Control how results are processed:

```yaml
results:
  limit: 10 # Maximum results to return
  deduplicate: true # Remove duplicate chunks
  include_score: false # Include similarity scores
  return_full_content: false # Return full docs vs chunks
```

## Common Use Cases

### Use Case 1: Project Documentation Assistant

Help users understand your project:

```yaml
rag:
  project_docs:
    description: "Project documentation and guides"
    docs:
      - "https://github.com/yourorg/yourproject/blob/main/README.md"
      - "https://yourproject.com/docs"
      - "./docs"
    strategies:
      - type: "chunked-embeddings"
        model: "text-embedding-3-small"
```

**Usage**:

```bash
don agent --tools=tools.yaml --rag=project_docs \
  "How do I install this project?"
```

### Use Case 2: API Documentation Assistant

Help developers use your API:

```yaml
rag:
  api_docs:
    description: "API reference and examples"
    docs:
      - "https://api.example.com/docs/v1"
      - "./api-docs"
      - "./examples/api"
    strategies:
      - type: "chunked-embeddings"
        model: "text-embedding-3-small"
        chunking:
          size: 800
          overlap: 150
```

**Usage**:

```bash
don agent --tools=tools.yaml --rag=api_docs \
  "Show me how to authenticate with the API"
```

### Use Case 3: Code Examples Assistant

Help users find relevant code examples:

```yaml
rag:
  code_examples:
    description: "Code examples and snippets"
    docs:
      - "./examples"
      - "./tests"
      - "./samples"
    strategies:
      - type: "chunked-embeddings"
        model: "text-embedding-3-small"
        chunking:
          size: 500
          overlap: 100
    results:
      return_full_content: true # Return complete code files
```

**Usage**:

```bash
don agent --tools=tools.yaml --rag=code_examples \
  "Show me an example of using the authentication module"
```

### Use Case 4: Multi-Source Knowledge Base

Combine multiple knowledge sources:

```yaml
rag:
  product_docs:
    description: "Product documentation"
    docs:
      - "https://example.com/docs"
      - "./docs/product"
    strategies:
      - type: "chunked-embeddings"
        model: "text-embedding-3-small"

  api_docs:
    description: "API documentation"
    docs:
      - "https://api.example.com/docs"
      - "./docs/api"
    strategies:
      - type: "chunked-embeddings"
        model: "text-embedding-3-small"

  security_docs:
    description: "Security guidelines"
    docs:
      - "./docs/security"
    strategies:
      - type: "chunked-embeddings"
        model: "text-embedding-3-small"
```

**Usage**:

```bash
# Use all sources
don agent --tools=tools.yaml \
  --rag=product_docs \
  --rag=api_docs \
  --rag=security_docs \
  "What are the security best practices for the API?"
```

## Interactive Mode

For ongoing conversations, omit the user prompt to enter interactive mode:

```bash
don agent --tools=tools.yaml --rag=docs
```

Then ask multiple questions:

```
> What is Don?
[Agent responds with information from docs]

> How do I configure it?
[Agent responds with configuration info]

> Show me an example
[Agent provides examples from docs]
```

## One-Shot Mode

For single questions, use the `--once` flag:

```bash
don agent --tools=tools.yaml --rag=docs --once \
  "What is Don?"
```

The agent will answer and exit immediately.

## Cache Management

### Cache Location

Documents are cached in platform-specific directories:

- **macOS**: `~/Library/Caches/don/rag/`
- **Linux**: `~/.cache/don/rag/`
- **Windows**: `%LOCALAPPDATA%\don\rag\`

### Custom Cache Directory

Override the cache location:

```bash
export DON_RAG_CACHE_DIR="/path/to/custom/cache"
don agent --tools=tools.yaml --rag=docs "Your question"
```

### Cache Behavior

- URLs are downloaded once and cached
- Cache is validated using ETag and Last-Modified headers
- Stale documents are automatically re-downloaded
- Local files are always scanned fresh (not cached)

### Clearing the Cache

To force re-download of all documents:

```bash
# macOS/Linux
rm -rf ~/.cache/don/rag/

# Or use custom location
rm -rf $DON_RAG_CACHE_DIR
```

## Tips and Best Practices

### 1. Start Small

Begin with a single RAG source and a few documents:

```yaml
rag:
  docs:
    description: "Getting started docs"
    docs:
      - "./README.md"
      - "./docs/getting-started.md"
    strategies:
      - type: "chunked-embeddings"
        model: "text-embedding-3-small"
```

### 2. Optimize Chunk Size

- **Larger chunks (1000-2000)**: Better for comprehensive context
- **Smaller chunks (500-800)**: Better for precise information retrieval
- **Overlap (100-200)**: Ensures context isn't lost at boundaries

### 3. Use Descriptive Names

Give your RAG sources clear, descriptive names:

```yaml
rag:
  getting_started_guide: # Clear purpose
    description: "Getting started documentation for new users"

  api_reference: # Clear purpose
    description: "Complete API reference documentation"
```

### 4. Organize by Topic

Create separate RAG sources for different topics:

```yaml
rag:
  installation:
    docs: ["./docs/install"]

  configuration:
    docs: ["./docs/config"]

  troubleshooting:
    docs: ["./docs/troubleshooting"]
```

Then use the relevant source:

```bash
don agent --tools=tools.yaml --rag=troubleshooting \
  "Why am I getting connection errors?"
```

### 5. Keep Documents Updated

- Use URLs for frequently updated documentation
- The cache will automatically refresh when documents change
- For local files, keep them in version control

### 6. Test Your Configuration

Test with known questions to verify the agent retrieves correct information:

```bash
don agent --tools=tools.yaml --rag=docs --once \
  "What is the installation command?"
```

## Troubleshooting

### Agent doesn't use RAG information

**Problem**: Agent answers without using your documents

**Solutions**:

- Verify RAG source name matches: `--rag=docs` must match `rag: docs:` in config
- Check documents are accessible (URLs return 200, local paths exist)
- Ensure documents contain relevant information
- Try more specific questions

### Documents not found

**Problem**: Error about missing documents

**Solutions**:

- Use absolute paths or paths relative to where you run the command
- Check file permissions
- Verify URLs are accessible: `curl -I <url>`

### Cache issues

**Problem**: Agent uses old document versions

**Solutions**:

- Clear the cache: `rm -rf ~/.cache/don/rag/`
- Check ETag/Last-Modified headers are set on your server
- Use local files instead of URLs for development

### API rate limits

**Problem**: OpenAI API rate limit errors

**Solutions**:

- Reduce chunk limit in strategies
- Reduce results limit
- Use smaller documents
- Implement retry logic in your workflow

## Current Status

RAG support is **fully implemented** and integrated with cagent's multi-agent framework.

**What works**:

- ✅ Document downloading and caching
- ✅ Local file and directory scanning
- ✅ Configuration parsing and conversion to cagent format
- ✅ CLI integration
- ✅ Document indexing with embeddings (via cagent)
- ✅ RAG tool creation in agent (via teamloader)
- ✅ Query-time retrieval (via cagent's RAG system)
- ✅ Multi-strategy retrieval (chunked-embeddings, BM25, hybrid)
- ✅ Result fusion (RRF, weighted, max)

See [RAG Architecture](arch-rag.md) for implementation details and roadmap.

## Next Steps

- **Learn more**: See [RAG Architecture](arch-rag.md) for technical details
- **Configure agent**: See [Agent Configuration](usage-agent-conf.md) for model setup
- **Examples**: Check `examples/agent_with_rag.yaml` for complete configuration
- **Troubleshooting**: See [Troubleshooting Guide](troubleshooting.md) for common issues

## Example: Complete Configuration

Here's a complete example combining everything:

```yaml
agent:
  # Model configuration
  models:
    - model: "gpt-4o"
      class: "openai"
      name: "openai"
      default: true
      api-key: "${OPENAI_API_KEY}"
      api-url: "https://api.openai.com/v1"

  # Orchestrator (optional, for complex workflows)
  orchestrator:
    model: "gpt-4o"
    class: "openai"
    api-key: "${OPENAI_API_KEY}"

  # RAG sources
  rag:
    # Product documentation
    product_docs:
      description: "Product documentation and user guides"
      docs:
        - "https://example.com/docs/getting-started.md"
        - "https://example.com/docs/user-guide.md"
        - "./docs/product"
      strategies:
        - type: "chunked-embeddings"
          model: "text-embedding-3-small"
          chunking:
            size: 1000
            overlap: 200
            respect_word_boundaries: true
          limit: 5
      results:
        limit: 10
        deduplicate: true
        include_score: false

    # API documentation
    api_docs:
      description: "API reference and examples"
      docs:
        - "https://api.example.com/docs/v1"
        - "./docs/api"
        - "./examples/api"
      strategies:
        - type: "chunked-embeddings"
          model: "text-embedding-3-small"
          chunking:
            size: 800
            overlap: 150
          limit: 3
        - type: "bm25"
          limit: 3
      results:
        limit: 5
        fusion:
          strategy: "rrf"
          k: 60
        deduplicate: true

# System prompts (optional)
prompts:
  system:
    - "You are a helpful assistant with access to documentation."
    - "Use the RAG tools to retrieve relevant information before answering."
    - "Cite the source documents when providing information."
```

**Usage**:

```bash
# Single source
don agent --tools=tools.yaml --rag=product_docs \
  "How do I get started?"

# Multiple sources
don agent --tools=tools.yaml \
  --rag=product_docs \
  --rag=api_docs \
  "Explain the authentication API"

# Interactive mode
don agent --tools=tools.yaml --rag=product_docs

# One-shot mode
don agent --tools=tools.yaml --rag=product_docs --once \
  "What is this product?"
```
