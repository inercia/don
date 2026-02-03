---
type: "always_apply"
---

# Architecture and Design Patterns for Don

## Project Structure

### Directory Organization
```
don/
├── cmd/                    # Command-line interface implementation
│   ├── root.go            # Root command and global flags
│   ├── agent.go           # Main agent command
│   ├── config.go          # Configuration management command
│   └── info.go            # Agent info command
├── pkg/                    # Core application packages
│   ├── agent/             # Agent mode implementation
│   ├── rag/               # RAG (Retrieval-Augmented Generation)
│   ├── common/            # Shared utilities and types
│   └── utils/             # Helper functions
├── docs/                   # Documentation
├── examples/              # Example configurations
├── tests/                 # Integration and E2E tests
├── build/                 # Build output directory
└── main.go                # Application entry point
```

### Package Responsibilities

#### `cmd/` Package
- Command-line interface implementation using Cobra
- Command definitions and flag parsing
- User interaction and output formatting
- Delegates business logic to `pkg/` packages

#### `pkg/agent/` Package
- Agent mode implementation
- LLM client integration (OpenAI, Anthropic, Ollama, etc.)
- Conversation management
- Tool execution orchestration using cagent framework
- MCPShell subprocess management for MCP tools

#### `pkg/rag/` Package
- Retrieval-Augmented Generation implementation
- Document loading and caching
- Embedding and search strategies
- Result fusion and deduplication

#### `pkg/common/` Package
- Shared types and interfaces
- Logging infrastructure
- Panic recovery

#### `pkg/utils/` Package
- Helper functions for file operations
- Path resolution and normalization
- Home directory detection (~/.don)

## Design Patterns

### Dependency Injection
- Pass dependencies (logger, config) as parameters to constructors
- Use constructor functions (New*) for complex types
- Avoid global state except for the global logger

### Interface-Based Design
- Define interfaces for pluggable components (ModelProvider, RAGStrategy)
- Use interfaces to enable testing with mocks
- Keep interfaces small and focused

### Factory Pattern
- Use factory functions for creating agents and RAG sources
- Factory functions handle initialization and validation

### Strategy Pattern
- Multiple RAG strategies (chunked-embeddings, BM25, hybrid)
- Model provider selection based on configuration
- Result fusion strategies (RRF, weighted, max)

## Architectural Principles

### Separation of Concerns
- Clear separation between CLI, business logic, and infrastructure
- Each package has a single, well-defined responsibility
- Avoid circular dependencies between packages

### Error Handling
- Errors are wrapped with context at each layer
- Use `fmt.Errorf` with `%w` for error wrapping
- Log errors at the point where they can be handled

### Logging Strategy
- Structured logging with levels (Debug, Info, Warn, Error)
- Logger passed as dependency, not accessed globally (except via GetLogger)

### Context Propagation
- Pass `context.Context` as first parameter for I/O operations
- Use context for cancellation and timeouts

### Configuration Management
- YAML-based configuration files (~/.don/agent.yaml)
- Environment variable substitution (${VAR_NAME})
- Default values for optional settings

## MCPShell Integration

### Tool Execution
- Don spawns MCPShell as a subprocess for MCP tools
- MCPShell configuration files define available tools
- Don converts LLM tool calls to MCP requests

### Configuration Files
- Don uses its own config (~/.don/agent.yaml) for LLM settings
- Don uses MCPShell format for tool definitions (--tools flag)

## Extension Points

### Adding New Model Providers
1. Implement the `ModelProvider` interface
2. Add provider-specific configuration
3. Register provider in model factory
4. Add tests for provider integration

### Adding New RAG Strategies
1. Implement the strategy interface
2. Add strategy-specific configuration
3. Register strategy in RAG factory
4. Add tests for strategy

## Testing Architecture

### Test Organization
- Unit tests in same package as source code (`*_test.go`)
- Integration tests in `tests/` directory
- Shell scripts for E2E testing

### Test Patterns
- Table-driven tests for multiple scenarios
- Test logger that discards output
- Mock implementations of interfaces

