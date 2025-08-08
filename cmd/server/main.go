package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	mcptools "github.com/frherrer/mcp-sail-operator/pkg/mcp"
)

var (
	kubeconfigPath string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "mcp-sail-operator",
		Short: "MCP server for Istio Sail Operator integration",
		Long: `MCP server that provides Claude with access to Istio Sail Operator resources 
and Kubernetes cluster information.`,
		Run: runServer,
	}

	// Add kubeconfig flag
	rootCmd.Flags().StringVar(&kubeconfigPath, "kubeconfig", "", 
		"Path to kubeconfig file (default: ~/.kube/config or KUBECONFIG env var)")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Command execution failed: %v", err)
	}
}

func runServer(cmd *cobra.Command, args []string) {
	// Initialize Kubernetes clients
	k8sClient, dynamicClient, err := initKubernetesClients(kubeconfigPath)
	if err != nil {
		log.Fatalf("Failed to initialize Kubernetes clients: %v", err)
	}

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "mcp-sail-operator",
		Version: "0.1.0",
	}, nil)

	// Register all MCP tools
	mcptools.RegisterAllTools(server, k8sClient, dynamicClient)

	// Start server using stdio transport
	log.Println("Starting MCP Sail Operator server...")
	if err := server.Run(context.Background(), mcp.NewStdioTransport()); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// initKubernetesClients creates both standard and dynamic Kubernetes clients using the specified or default kubeconfig
func initKubernetesClients(kubeconfigPath string) (*kubernetes.Clientset, dynamic.Interface, error) {
	var config *rest.Config
	var err error

	// Try to use in-cluster config first
	config, err = rest.InClusterConfig()
	if err != nil {
		// Determine kubeconfig path
		configPath := kubeconfigPath
		if configPath == "" {
			// Use default kubeconfig path
			configPath = clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
			if envPath := os.Getenv("KUBECONFIG"); envPath != "" {
				configPath = envPath
			}
		}

		log.Printf("Using kubeconfig: %s", configPath)

		// Build config from kubeconfig file
		config, err = clientcmd.BuildConfigFromFlags("", configPath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to build kubeconfig from %s: %w", configPath, err)
		}
	} else {
		log.Println("Using in-cluster configuration")
	}

	// Create the standard clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Create the dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return clientset, dynamicClient, nil
}

