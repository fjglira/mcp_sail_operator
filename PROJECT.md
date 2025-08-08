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

### MCP Tools Implemented

**Kubernetes Tools:**
1. `test_k8s_connection` - Test cluster connectivity and version
2. `list_namespaces` - List all namespaces in the cluster
3. `get_namespace_details` - Get detailed namespace information
4. `list_pods` - List pods with status, ready state, and filtering options
5. `list_services` - List services with networking and port information
6. `list_deployments` - List deployments with replica and status information
7. `list_configmaps` - List configmaps with data keys and filtering options

**Sail Operator Tools:**
4. `list_sailoperator_resources` - List all Sail Operator CRD resources
5. `get_istio_status` - Get detailed status of Istio installations
6. `check_sailoperator_health` - Comprehensive health checks

### MCP Tools Planned

1. `get_mesh_config` - Retrieve mesh configuration and settings
2. `validate_config` - Validate Istio resource configurations
3. `get_revision_info` - Get information about specific Istio revisions
4. `list_workloads` - List workloads using the service mesh
5. `get_security_policies` - Retrieve security policies and configurations
6. `list_pods` - List pods in namespaces
7. `get_service_status` - Get service mesh sidecar injection status

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
  pkg/handlers/
  ├── k8s/          (Basic K8s operations)
  ├── istio/        (Mesh operations)
  └── sailoperator/ (CRD handlers)
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
- [x] Complete Kubernetes resource tools: connectivity, namespaces, pods, services, deployments, configmaps
- [x] Sail Operator CRD handlers: list_sailoperator_resources, get_istio_status, check_sailoperator_health
- [x] Build system with Makefile
- [x] Refactored codebase into organized package structure with dynamic client support
- [x] Comprehensive health checking and status monitoring for Istio mesh
- [ ] Istio mesh status and health check tools
- [ ] Configuration validation and troubleshooting tools
- [ ] Testing and validation with live cluster

## Contributing

This project follows the company's development practices and is designed to be a learning exercise while creating valuable tooling for the Istio and Kubernetes community.

## License

This project is licensed under the MIT License - see the LICENSE file for details.