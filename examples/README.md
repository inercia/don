# Don Examples

This directory contains example configuration files for Don.

## Files

### `config-with-rag.yaml`

Example agent configuration with RAG (Retrieval-Augmented Generation) support.
This demonstrates how to configure document sources and retrieval strategies.

## Usage

Copy the example you want to use as a starting point:

```bash
cp config-with-rag.yaml ~/.don/agent.yaml
```

Then edit the file to customize the settings for your needs.

## Tool Configuration

Don uses MCPShell's YAML format for tool configuration. You can find example
tool configurations in the MCPShell repository:

```bash
# Clone MCPShell for example tools
git clone https://github.com/inercia/MCPShell
ls MCPShell/examples/
```

Or create your own tools configuration. Here's a minimal example:

```yaml
mcp:
  description: "Basic file tools"
  run:
    shell: bash
  tools:
    - name: "list_files"
      description: "List files in a directory"
      run:
        command: "ls -la {{ .directory }}"
      params:
        directory:
          type: string
          description: "Directory to list"
          required: true
```

Save this to a file (e.g., `tools.yaml`) and use it with Don:

```bash
don --tools=tools.yaml "List the files in my home directory"
```

