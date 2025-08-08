# MCP Sail Operator Server

🚀 **The first Model Context Protocol (MCP) server for Kubernetes and Istio integration!**

A comprehensive MCP server that provides Claude with read-only access to Istio Sail Operator resources and Kubernetes cluster information. Features both direct CLI commands and natural language interaction through Claude Code.

## 🎯 Key Features

- **🔧 Dual Interface**: Direct CLI commands + Natural language queries through Claude
- **🔒 Security-First**: Read-only operations for safe cluster monitoring
- **🕸️ Mesh Analysis**: Complete Istio service mesh monitoring and workload analysis
- **📊 Health Monitoring**: Comprehensive cluster and mesh health checking
- **⚡ Real-time**: Live cluster status and resource information
- **🎯 Sail Operator**: Native support for cluster-scoped Istio CRDs

## 🛠️ Complete Tool Suite

### 🔧 CLI Commands (Direct Usage)
```bash
# Pod and Resource Management
./mcp-sail-operator pods                                    # List all pods
./mcp-sail-operator pods --namespace istio-system           # Namespace-specific
./mcp-sail-operator logs istiod-abc123 -n istio-system -l 50  # Pod logs

# Health and Status Monitoring  
./mcp-sail-operator health                                  # Comprehensive health check
./mcp-sail-operator status                                  # Istio installation status
```

### 🤖 MCP Tools (Natural Language)

#### Kubernetes Operations (9 tools)
- `test_k8s_connection` - Test cluster connectivity and version information
- `list_namespaces` - List all namespaces with metadata
- `get_namespace_details` - Detailed namespace information with labels/annotations
- `list_pods` - Pod listing with status, ready state, restarts, age + filtering
- `list_services` - Service listing with types, IPs, ports + filtering  
- `list_deployments` - Deployment status with replica counts and strategies
- `list_configmaps` - ConfigMap listing with data counts and keys
- `get_pod_logs` - Pod log retrieval with container selection and line limits
- `check_mesh_workloads` - **Mesh workload analysis with sidecar injection status**

#### Sail Operator Integration (3 tools)
- `list_sailoperator_resources` - List cluster-scoped CRDs (Istio, IstioRevision, IstioCNI, ZTunnel)
- `get_istio_status` - Detailed Istio installation status with revisions and conditions
- `check_sailoperator_health` - Comprehensive health checks for all Sail Operator components

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
         "command": "/path/to/your/mcp-sail-operator/mcp-sail-operator",
         "args": [],
         "env": {}
       }
     }
   }
   ```
   
   **Important Notes:**
   - Replace `/path/to/your/mcp-sail-operator/` with the actual path to your project directory
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

## 🗣️ Natural Language Examples

Once configured with Claude Code, you can interact naturally:

### Cluster Monitoring
```
"Are my pods running correctly?"
"Show me what's happening in the istio-system namespace"
"Is my Kubernetes cluster healthy?"
"What services are running in the default namespace?"
```

### Mesh Analysis  
```
"Are all my workloads properly injected with Istio sidecars?"
"Is my Istio installation healthy?"
"Check if there are any issues with my service mesh"
"Show me the status of my Istio components"
```

### Troubleshooting
```
"Can you check the logs of the istiod pod for any errors?"
"Are there any pods that failed to start?"
"What's the status of my Sail Operator deployment?"
"Show me any configuration issues in my mesh"
```

### Resource Discovery
```
"What Istio CRDs are installed in my cluster?"
"List all the pods that have Istio sidecars"
"Show me the health of all Sail Operator components"
```

## 🎉 Project Status: COMPLETED

✅ **All objectives successfully achieved!**

This project was completed as part of a learning day initiative to create the first MCP server for Kubernetes/Istio:

### ✅ Core Implementation
- **MCP Server**: Full stdio transport implementation with Go SDK
- **Kubernetes Integration**: Complete client-go integration with kubeconfig support  
- **12 Working Tools**: Comprehensive Kubernetes and Sail Operator tool suite
- **Claude Code Integration**: Fully configured and documented setup

### ✅ Advanced Features  
- **Dual Interface**: Both CLI commands and natural language interaction
- **Security-First Design**: Read-only operations for safe cluster monitoring
- **Mesh Analysis**: Complete workload sidecar injection status checking
- **Health Monitoring**: Comprehensive Sail Operator and Istio health checks
- **Cluster-Scoped Resources**: Correct handling of Istio CRDs (Istio, IstioRevision, etc.)

### ✅ Code Quality
- **Organized Architecture**: Clean package separation (k8s, sailoperator, mesh handlers)
- **Error Handling**: Robust error handling and user-friendly output formatting
- **Documentation**: Complete docs with examples and usage patterns
- **Build System**: Full Makefile with build, test, and development targets

### 🚀 Ready for Production
This MCP server is fully functional and ready for real-world Kubernetes/Istio cluster monitoring!

## 🤝 Development Collaboration

This project was developed in collaboration with **Claude (Anthropic)** as part of a learning day initiative, showcasing the power of AI-assisted software development for cloud-native infrastructure tooling.

### Development Process
- **Human-AI Collaboration**: Leveraged Claude's expertise in Go, Kubernetes, and MCP protocols
- **Iterative Development**: Real-time code review, architecture guidance, and best practices
- **Knowledge Transfer**: Learning Kubernetes operators, MCP ecosystem, and service mesh concepts
- **Quality Assurance**: Code organization, error handling, and documentation standards

The collaboration demonstrates how AI can accelerate development of complex infrastructure tools while maintaining high code quality and comprehensive documentation.

## Contributing

This project follows standard Go development practices and is designed to be a valuable tool for the Istio and Kubernetes community.

## License

MIT License - see LICENSE file for details

---

*🤖 Developed with [Claude Code](https://claude.ai/code) - AI-assisted development for cloud-native infrastructure*
