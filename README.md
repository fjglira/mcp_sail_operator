# MCP Sail Operator Server

An MCP (Model Context Protocol) server that provides Claude with access to Istio Sail Operator resources and Kubernetes cluster information. This is a personal project to learn more about MCP server and AI integration with tools.

## Features

- **Kubernetes Connectivity**: Test cluster connection and list resources
- **Namespace Management**: List and explore cluster namespaces  
- **Istio Integration**: (Coming soon) Access to Istio mesh status and configuration
- **Sail Operator CRDs**: (Coming soon) Query Istio, IstioRevision, IstioCNI resources

## Available MCP Tools

- `test_k8s_connection` - Test connectivity to the Kubernetes cluster
- `list_namespaces` - List all namespaces in the cluster
- `get_namespace_details` - Get detailed information about a specific namespace

## Prerequisites

- Go 1.21 or later
- Access to a Kubernetes cluster
- Valid kubeconfig file (default: `~/.kube/config`)

## Quick Start

1. **Build the server:**
   ```bash
   make build
   ```

2. **Run the server:**
   ```bash
   make run
   ```

3. **Configure with Claude Code:**
   
   Create or update your Claude Code settings file at `~/.claude/settings.local.json`:
   ```json
   {
     "mcpServers": {
       "sail-operator": {
         "command": "/Users/frherrer/Documents/repos/mcp_sail_operator/mcp-sail-operator",
         "args": [],
         "env": {}
       }
     }
   }
   ```
   
   **Important Notes:**
   - Use the absolute path to your compiled `mcp-sail-operator` binary
   - Ensure the binary is built with `make build` before configuring
   - Claude Code will restart the MCP server automatically when needed
   - The server uses your default kubeconfig (`~/.kube/config`) or `KUBECONFIG` environment variable

## Development

### Building
```bash
# Build binary
make build

# Clean artifacts  
make clean

# Format code
make fmt

# Tidy dependencies
make tidy
```

### Testing
```bash
# Run tests
make test

# Test Kubernetes connectivity
./mcp-sail-operator
```

## Configuration

The server uses standard Kubernetes client configuration:

- Default kubeconfig: `~/.kube/config`
- Environment variable: `KUBECONFIG=/path/to/config`
- In-cluster config when running as a pod

## Using with Claude Code

Once configured, you can use the MCP tools in Claude Code conversations:

```
# Test your cluster connection
Can you test my Kubernetes cluster connection?

# List namespaces
Show me all namespaces in my cluster

# Get namespace details
Give me detailed information about the default namespace
```

## Project Status

This project is under active development as part of a learning day initiative. Current status:

- ✅ Basic MCP server implementation with stdio transport
- ✅ Kubernetes client integration with kubeconfig support
- ✅ Three working MCP tools: connectivity, namespace listing, and namespace details
- ✅ Claude Code integration and local setup instructions
- 🚧 Sail Operator CRD handlers (Istio, IstioRevision, IstioCNI, Ztunnel)
- 🚧 Istio mesh status and health check tools
- ⏳ Advanced troubleshooting and configuration validation features

## Contributing

This project follows standard Go development practices and is designed to be a valuable tool for the Istio and Kubernetes community.

## License

MIT License - see LICENSE file for details
