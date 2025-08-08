# MCP Sail Operator Server

## Overview

This project implements a Model Context Protocol (MCP) server that provides Claude with access to Istio Sail Operator resources and Kubernetes cluster information. The MCP server enables Claude to understand, monitor, and help troubleshoot Istio service mesh deployments managed by the Sail Operator.

## Background

The Sail Operator is an open-source Kubernetes operator that manages the lifecycle of Istio service mesh installations. It provides custom resources like `Istio`, `IstioRevision`, `IstioCNI`, and `Ztunnel` to deploy and manage Istio control plane components.

Currently, there are no existing MCP servers for Kubernetes or Istio in the official MCP ecosystem, making this project a valuable contribution to both the company and the broader community.

## Objectives

1. Create the first MCP server for Kubernetes and Istio integration
2. Provide Claude with real-time access to Sail Operator managed resources
3. Enable intelligent troubleshooting and configuration assistance
4. Improve developer experience when working with Istio service mesh

## Features

### Planned Capabilities

- **Resource Discovery**: List and describe Istio, IstioRevision, IstioCNI, and Ztunnel resources
- **Status Monitoring**: Real-time health checks and status reporting
- **Configuration Analysis**: Validate and suggest improvements for Istio configurations
- **Troubleshooting**: Identify common issues and provide resolution guidance
- **Version Management**: Check version compatibility and upgrade paths
- **Security Analysis**: Review security policies and configurations

### MCP Tools to Implement

1. `list_istio_resources` - List all Istio-related custom resources
2. `get_istio_status` - Get detailed status of specific Istio installations
3. `get_mesh_config` - Retrieve mesh configuration and settings
4. `check_istio_health` - Perform health checks on Istio components
5. `validate_config` - Validate Istio resource configurations
6. `get_revision_info` - Get information about specific Istio revisions
7. `list_workloads` - List workloads using the service mesh
8. `get_security_policies` - Retrieve security policies and configurations

## Technology Stack

- **Language**: Go
- **MCP SDK**: Go MCP SDK
- **Kubernetes Client**: client-go library
- **Custom Resources**: Generated clients for Sail Operator CRDs

## Architecture

```
Claude Client
     |
     | MCP Protocol
     |
MCP Sail Operator Server
     |
     | Kubernetes API
     |
Kubernetes Cluster
     |
Sail Operator + Istio Resources
```

## Getting Started

### Prerequisites

- Go 1.21 or later
- Access to a Kubernetes cluster with Sail Operator deployed
- Valid kubeconfig file

### Installation

1. Clone the repository
2. Install dependencies: `go mod tidy`
3. Build the server: `go build -o mcp-sail-operator ./cmd/server`
4. Run the server: `./mcp-sail-operator`

### Configuration

The server will use the default kubeconfig location (`~/.kube/config`) or respect the `KUBECONFIG` environment variable.

#### Claude Code Integration

To use this MCP server with Claude Code, create or update `~/.claude/settings.local.json`:

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

**Key Points:**
- Use absolute path to the compiled binary
- Build the binary first with `make build`
- Claude Code will manage the MCP server lifecycle
- Server uses your default Kubernetes configuration

## Development Status

This project is in active development as part of a learning day initiative. Current status:

- [x] Project planning and research
- [x] Initial project structure
- [x] Go module initialization with MCP SDK and Kubernetes dependencies
- [x] Basic MCP server implementation with stdio transport
- [x] Kubernetes client integration with kubeconfig support
- [x] Basic tools: test_k8s_connection, list_namespaces, and get_namespace_details
- [x] Build system with Makefile
- [ ] Sail Operator CRD handlers (Istio, IstioRevision, IstioCNI, Ztunnel)
- [ ] Istio mesh status and health check tools
- [ ] Configuration validation and troubleshooting tools
- [ ] Testing and validation with live cluster

## Contributing

This project follows the company's development practices and is designed to be a learning exercise while creating valuable tooling for the Istio and Kubernetes community.

## License

This project is licensed under the MIT License - see the LICENSE file for details.