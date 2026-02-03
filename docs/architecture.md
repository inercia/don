# RAG Architecture

This document describes the Retrieval-Augmented Generation (RAG) architecture in
Don, including implementation status, design decisions, and usage.

## Overview

RAG support allows Don agents to retrieve relevant information from documents
(URLs, local files, or directories) to answer questions. The implementation provides a
document processing foundation that integrates with the
[cagent](https://github.com/docker/cagent) library's RAG system.

## Architecture

### Components

```
┌─────────────────────────────────────────────────────────────┐
│                         CLI Layer                            │
│  cmd/agent.go: --rag flag, config merging                   │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                      Agent Layer                             │
│  pkg/agent/agent.go: RAG source processing                  │
│  pkg/agent/rag.go: Document processing                      │
│  pkg/agent/cagent_runtime.go: cagent integration           │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                 Document Processing Layer                    │
│  pkg/rag/downloader.go: HTTP downloads with caching         │
│  pkg/rag/scanner.go: Local file/directory scanning         │
│  pkg/rag/cache.go: Platform-specific cache management      │
└─────────────────────────────────────────────────────────────┘
```

### Package Structure

#### `pkg/rag/` - Document Processing

Core document processing functionality:

- **types.go**: Core types and interfaces
  - `DocumentSource`: Represents a document (URL or local path)
  - `Downloader`: Interface for downloading documents
  - `Scanner`: Interface for scanning local files
  - `DownloaderConfig`: Configuration for document processing

- **downloader.go**: HTTP document downloader
  - Intelligent caching with ETag/Last-Modified headers
  - Content-type validation (text files only)
  - Size limits and timeout handling
  - Automatic re-download when documents are stale

- **scanner.go**: File and directory scanner
  - Recursive directory scanning
  - Text file detection based on extension
  - Path validation and security checks

- **cache.go**: Cache management
  - Platform-specific cache directories:
    - macOS: `~/Library/Caches/don/rag/`
    - Linux: `~/.cache/don/rag/`
    - Windows: `%LOCALAPPDATA%\don\rag\`
  - Override with `DON_RAG_CACHE_DIR` environment variable
  - Path traversal prevention
  - Metadata serialization

- **errors.go**: Custom error types for RAG operations

#### `pkg/agent/` - RAG Integration

Integration with agent functionality:

- **config_file.go**: RAG configuration structures
  - `RAGSourceConfig`: Top-level RAG source definition
  - `RAGStrategyConfig`: Retrieval strategy (chunked-embeddings, BM25)
  - `RAGChunkingConfig`: Document chunking parameters
  - `RAGFusionConfig`: Multi-strategy result fusion
  - `RAGResultsConfig`: Result processing options

- **rag.go**: Document processing logic
  - `ProcessRAGSources()`: Downloads URLs and scans local files
  - Converts all document sources to local paths
  - Returns processed configuration ready for indexing

- **agent.go**: Agent initialization
  - Added `RAGSources []string` and `RAGConfig map[string]RAGSourceConfig`
  - Processes RAG sources before agent starts
  - Documents downloaded and scanned during initialization

- **cagent_runtime.go**: cagent integration
  - Uses teamloader to load team with full RAG support
  - Generates cagent-compatible YAML configuration
  - RAG managers created automatically by cagent

#### `cmd/` - CLI Integration

Command-line interface:

- **root.go**: Flag variables
  - `agentRAGSources []string`: RAG source names to enable

- **agent.go**: Agent command
  - `--rag` flag: Specify RAG sources from config (repeatable)
  - `buildAgentConfig()`: Merges CLI flags with config file
  - Validates RAG sources exist in agent config

## Configuration

### Agent Configuration File

RAG sources are defined in the agent configuration file (e.g.,
`~/.don/agent.yaml`):

```yaml
agent:
  models:
    - model: "gpt-4o"
      class: "openai"
      default: true
      api-key: "${OPENAI_API_KEY}"

  rag:
    # Define RAG sources
    docs:
      description: "Project documentation"

      # Documents to index (URLs and local paths)
      docs:
        - "https://example.com/docs/getting-started.md"
        - "https://example.com/docs/api-reference.md"
        - "./docs" # Local directory
        - "./README.md" # Local file

      # Retrieval strategies
      strategies:
        - type: "chunked-embeddings"
          model: "text-embedding-3-small"
          chunking:
            size: 1000
            overlap: 200
            respect_word_boundaries: true
          limit: 5

      # Result configuration
      results:
        limit: 10
        deduplicate: true
        include_score: false
```

### Configuration Options

**RAG Source Fields**:

- `description`: Human-readable description of the knowledge source
- `docs`: List of URLs or local paths (files or directories)
- `strategies`: List of retrieval strategies to use
- `results`: Result processing configuration

**Strategy Types**:

- `chunked-embeddings`: Vector similarity search with embeddings
- `bm25`: BM25 keyword-based search

**Chunking Options**:

- `size`: Chunk size in characters (default: 1000)
- `overlap`: Overlap between chunks in characters (default: 200)
- `respect_word_boundaries`: Don't split words (default: true)

**Results Options**:

- `limit`: Maximum number of results to return
- `fusion`: Multi-strategy result fusion (rrf, weighted, max)
- `deduplicate`: Remove duplicate results
- `include_score`: Include similarity scores
- `return_full_content`: Return full documents instead of chunks

## Usage

### Command-Line

Enable RAG sources when running the agent:

```bash
# Single RAG source
don agent --tools=tools.yaml --rag=docs "How do I get started?"

# Multiple RAG sources
don agent --tools=tools.yaml --rag=docs --rag=api "Explain the API"
```

### Example Workflow

1. **Create agent configuration** with RAG sources:

   ```bash
   cat > ~/.don/agent.yaml << EOF
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
   EOF
   ```

2. **Run agent with RAG enabled**:

   ```bash
   don agent --tools=tools.yaml --rag=docs "What is this project about?"
   ```

3. **Documents are processed**:
   - URLs are downloaded to cache
   - Local files and directories are scanned
   - All documents converted to local paths
   - Ready for indexing by cagent's RAG system

## Implementation Status

### ✅ Phase 1: Document Processing (COMPLETE)

**Status**: Fully implemented and tested

**Features**:

- HTTP document downloader with intelligent caching
- File scanner for local files and directories
- Platform-specific cache directories
- Comprehensive security features
- All tests passing

**Files**:

- `pkg/rag/types.go` - Core types and interfaces
- `pkg/rag/errors.go` - Custom error types
- `pkg/rag/cache.go` - Cache management
- `pkg/rag/downloader.go` - HTTP downloader (297 lines)
- `pkg/rag/scanner.go` - File scanner (159 lines)
- `pkg/rag/cache_test.go` - Test suite

### ✅ Phase 2: RAG Integration (COMPLETE)

**Status**: Fully implemented using cagent's teamloader

**Features**:

- ✅ RAG configuration parsing from YAML
- ✅ Document downloading from URLs with caching
- ✅ Local file and directory scanning
- ✅ Document path resolution and validation
- ✅ All processed documents available as local paths
- ✅ Config conversion to cagent v2 YAML format
- ✅ RAG manager creation via teamloader
- ✅ Document indexing with embeddings (via cagent)
- ✅ RAG tool creation and registration (via cagent)
- ✅ Manager lifecycle (automatic via teamloader)
- ✅ Multi-strategy retrieval support
- ✅ Result fusion (RRF, weighted, max)

**Implementation Approach**:

Don uses cagent's `teamloader.Load()` to create teams with full RAG support:

1. **Config Conversion** (`pkg/agent/config_yaml.go`):
   - Converts Don config to cagent v2 YAML format
   - Maps models, RAG sources, and strategies
   - Generates temporary config file

2. **Team Loading** (`pkg/agent/cagent_runtime.go`):
   - Calls `teamloader.Load()` with generated config
   - Teamloader handles all RAG setup automatically
   - RAG managers, tools, and indexing done by cagent

3. **MCP Tools** (subprocess):
   - MCP server started as subprocess via `command: "don mcp --tools <file>"`
   - Clean separation of concerns
   - Aligns with cagent's design

**Files**:

- `pkg/agent/config_file.go` - RAG configuration structures
- `pkg/agent/config_yaml.go` - Config conversion to cagent format (NEW - 317 lines)
- `pkg/agent/rag.go` - Document processing logic
- `pkg/agent/agent.go` - RAG source processing
- `pkg/agent/cagent_runtime.go` - Teamloader integration (REFACTORED - 183 lines)
- `go.mod` - Upgraded cagent to v1.9.26

### ✅ Phase 3: CLI Integration (COMPLETE)

**Status**: Fully implemented

**Features**:

- `--rag` flag for specifying RAG sources
- Configuration merging in `buildAgentConfig()`
- Example configuration with comprehensive documentation
- Updated README with RAG support note

**Files**:

- `cmd/root.go` - RAG flag variable
- `cmd/agent.go` - RAG flag and configuration merging
- `examples/agent_with_rag.yaml` - Example configuration
- `README.md` - RAG documentation

## Testing

### Unit Tests

All unit tests pass:

```bash
$ go test ./pkg/rag/...
=== RUN   TestGetCacheDir
--- PASS: TestGetCacheDir (0.00s)
=== RUN   TestHashURL
--- PASS: TestHashURL (0.00s)
=== RUN   TestSaveAndLoadMetadata
--- PASS: TestSaveAndLoadMetadata (0.00s)
=== RUN   TestValidatePath
--- PASS: TestValidatePath (0.00s)
PASS
ok      github.com/inercia/Don/pkg/rag     (cached)
```

### Integration Tests

Comprehensive integration test suite:

**Test Documents** (`tests/agent/rag_docs/`):

- `product_info.txt` - Product information and features
- `api_reference.txt` - API documentation
- `security_guide.txt` - Security guidelines

**Test Configuration** (`tests/agent/tools/test_agent_rag.yaml`):

- Agent with two RAG sources (product_docs, security_docs)
- Uses OpenAI GPT-4o-mini model
- Chunked-embeddings strategy

**Test Script** (`tests/agent/test_agent_rag.sh`):

- Tests single RAG source retrieval
- Tests multiple RAG sources simultaneously
- Verifies agent responses include RAG information
- Automatically skips if OPENAI_API_KEY not set

**Running Tests**:

```bash
# Set OpenAI API key
export OPENAI_API_KEY="sk-..."

# Run RAG test
./tests/agent/test_agent_rag.sh

# Or run all tests
make test-e2e
```

## Security

### Path Traversal Prevention

All paths are validated to prevent directory traversal attacks:

```go
func ValidatePath(path string) error {
    // Check for path traversal before and after cleaning
    if strings.Contains(path, "..") {
        return ErrInvalidPath
    }
    cleaned := filepath.Clean(path)
    if strings.Contains(cleaned, "..") {
        return ErrInvalidPath
    }
    return nil
}
```

### URL Validation

Only HTTP/HTTPS URLs are allowed:

```go
if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
    return "", fmt.Errorf("invalid URL scheme: %s", url)
}
```

### Content-Type Validation

Only text-based content types are downloaded:

```go
contentType := resp.Header.Get("Content-Type")
if !strings.HasPrefix(contentType, "text/") &&
   !strings.Contains(contentType, "json") &&
   !strings.Contains(contentType, "yaml") {
    return "", fmt.Errorf("unsupported content type: %s", contentType)
}
```

### Size Limits

Documents are limited to prevent memory exhaustion:

```go
const DefaultMaxSize = 10 * 1024 * 1024 // 10MB
```

## Next Steps for Full RAG Support

To complete the RAG integration, choose one of two approaches:

### Option A: Use cagent's teamloader (Recommended)

**Approach**: Replace custom agent creation with cagent's teamloader

**Pros**:

- Full RAG support out of the box
- Follows cagent's intended usage pattern
- Less custom code to maintain
- Automatic updates when cagent changes

**Cons**:

- Major refactoring of `pkg/agent/cagent_runtime.go`
- Lose some control over agent creation
- Need to adapt config format to cagent's expectations

**Implementation Steps**:

1. Convert our config to cagent's `latest.Config` format
2. Replace `CreateCagentRuntime()` to use `teamloader.Load()`
3. Pass processed RAG documents to the teamloader
4. Test with example RAG configurations

**Estimated Effort**: 1-2 days

### Option B: Manual RAG Manager Creation

**Approach**: Manually create RAG managers and tools

**Pros**:

- Keep current agent creation approach
- More control over RAG setup
- Can customize RAG behavior

**Cons**:

- Complex implementation
- Tightly coupled to cagent internals
- May break with cagent updates
- More code to maintain

**Implementation Steps**:

1. Study cagent's `pkg/rag/builder.go` and `pkg/teamloader/teamloader.go`
2. Implement conversion from our config to cagent's RAG config format
3. Create `ManagersBuildConfig` with required dependencies
4. Call `rag.NewManagers()` to create managers
5. Create RAG tools using `builtin.NewRAGTool()`
6. Add RAG tools to agent's toolset
7. Pass RAG managers to team via `team.WithRAGManagers()`

**Estimated Effort**: 3-5 days

### Recommendation

**Use Option A (teamloader)** for the following reasons:

1. It's the intended way to use cagent's RAG functionality
2. Less code to maintain and debug
3. Automatic compatibility with future cagent updates
4. The refactoring is manageable and well-scoped

## Design Decisions

### Why Platform-Specific Cache Directories?

Following OS conventions for cache storage:

- Better integration with OS cache management
- Users expect caches in standard locations
- Easier to find and clear caches
- Respects OS-specific cleanup policies

### Why Separate Document Processing Layer?

Separation of concerns:

- Document processing is independent of RAG indexing
- Can be tested in isolation
- Reusable for other purposes
- Clear interface boundaries

### Why Support Both URLs and Local Paths?

Flexibility for different use cases:

- URLs: Public documentation, API references
- Local paths: Private docs, code repositories
- Directories: Scan entire doc trees
- Files: Specific documents

### Why Intelligent Caching?

Performance and cost optimization:

- Avoid re-downloading unchanged documents
- Reduce network traffic
- Faster agent startup
- Lower API costs for embeddings

## Examples

See `examples/agent_with_rag.yaml` for a complete example configuration with:

- Three RAG sources (project_docs, api_docs, code_examples)
- Multiple retrieval strategies
- Result fusion configuration
- Comprehensive inline documentation

## References

- [cagent RAG PR #843](https://github.com/docker/cagent/pull/843) - Original RAG
  implementation
- [cagent Documentation](https://github.com/docker/cagent) - cagent library
- [Model Context Protocol](https://modelcontextprotocol.io/) - MCP specification
