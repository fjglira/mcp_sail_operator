package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	// Initialize Kubernetes client
	k8sClient, err := initKubernetesClient(kubeconfigPath)
	if err != nil {
		log.Fatalf("Failed to initialize Kubernetes client: %v", err)
	}

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "mcp-sail-operator",
		Version: "0.1.0",
	}, nil)

	// Register tools
	registerTools(server, k8sClient)

	// Start server using stdio transport
	log.Println("Starting MCP Sail Operator server...")
	if err := server.Run(context.Background(), mcp.NewStdioTransport()); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// initKubernetesClient creates a Kubernetes client using the specified or default kubeconfig
func initKubernetesClient(kubeconfigPath string) (*kubernetes.Clientset, error) {
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
			return nil, fmt.Errorf("failed to build kubeconfig from %s: %w", configPath, err)
		}
	} else {
		log.Println("Using in-cluster configuration")
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return clientset, nil
}

// TestConnectionParams represents parameters for the test connection tool
type TestConnectionParams struct{}

// TestConnectionResult represents the result of testing Kubernetes connection
type TestConnectionResult struct {
	Status            string `json:"status"`
	KubernetesVersion string `json:"kubernetes_version,omitempty"`
	ServerVersion     string `json:"server_version,omitempty"`
	Error             string `json:"error,omitempty"`
}

// ListNamespacesParams represents parameters for listing namespaces
type ListNamespacesParams struct{}

// ListNamespacesResult represents the result of listing namespaces
type ListNamespacesResult struct {
	Status     string   `json:"status"`
	Namespaces []string `json:"namespaces,omitempty"`
	Count      int      `json:"count,omitempty"`
	Error      string   `json:"error,omitempty"`
}

// GetNamespaceDetailsParams represents parameters for getting namespace details
type GetNamespaceDetailsParams struct {
	Namespace string `json:"namespace,omitempty"`
}

// NamespaceDetail represents detailed information about a namespace
type NamespaceDetail struct {
	Name        string            `json:"name"`
	Status      string            `json:"status"`
	CreatedAt   string            `json:"created_at"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// GetNamespaceDetailsResult represents the result of getting namespace details
type GetNamespaceDetailsResult struct {
	Status     string            `json:"status"`
	Namespaces []NamespaceDetail `json:"namespaces,omitempty"`
	Error      string            `json:"error,omitempty"`
}

// testK8sConnection tool handler
func testK8sConnection(k8sClient *kubernetes.Clientset) func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[TestConnectionParams]) (*mcp.CallToolResultFor[TestConnectionResult], error) {
	return func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[TestConnectionParams]) (*mcp.CallToolResultFor[TestConnectionResult], error) {
		// Try to get cluster version
		version, err := k8sClient.Discovery().ServerVersion()
		if err != nil {
			return &mcp.CallToolResultFor[TestConnectionResult]{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error connecting to Kubernetes: %v", err),
				}},
			}, nil
		}

		result := TestConnectionResult{
			Status:            "connected",
			KubernetesVersion: version.String(),
			ServerVersion:     version.GitVersion,
		}

		return &mcp.CallToolResultFor[TestConnectionResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Successfully connected to Kubernetes cluster.\nVersion: %s\nServer: %s", 
					result.KubernetesVersion, result.ServerVersion),
			}},
		}, nil
	}
}

// listNamespaces tool handler
func listNamespaces(k8sClient *kubernetes.Clientset) func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[ListNamespacesParams]) (*mcp.CallToolResultFor[ListNamespacesResult], error) {
	return func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[ListNamespacesParams]) (*mcp.CallToolResultFor[ListNamespacesResult], error) {
		namespaces, err := k8sClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			return &mcp.CallToolResultFor[ListNamespacesResult]{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error listing namespaces: %v", err),
				}},
			}, nil
		}

		var nsNames []string
		for _, ns := range namespaces.Items {
			nsNames = append(nsNames, ns.Name)
		}

		result := ListNamespacesResult{
			Status:     "success",
			Namespaces: nsNames,
			Count:      len(nsNames),
		}

		return &mcp.CallToolResultFor[ListNamespacesResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Found %d namespaces: %v", result.Count, result.Namespaces),
			}},
		}, nil
	}
}

// getNamespaceDetails tool handler
func getNamespaceDetails(k8sClient *kubernetes.Clientset) func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[GetNamespaceDetailsParams]) (*mcp.CallToolResultFor[GetNamespaceDetailsResult], error) {
	return func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[GetNamespaceDetailsParams]) (*mcp.CallToolResultFor[GetNamespaceDetailsResult], error) {
		var namespaces []NamespaceDetail
		
		if params.Arguments.Namespace != "" {
			// Get specific namespace
			ns, err := k8sClient.CoreV1().Namespaces().Get(ctx, params.Arguments.Namespace, metav1.GetOptions{})
			if err != nil {
				return &mcp.CallToolResultFor[GetNamespaceDetailsResult]{
					Content: []mcp.Content{&mcp.TextContent{
						Text: fmt.Sprintf("Error getting namespace %s: %v", params.Arguments.Namespace, err),
					}},
				}, nil
			}
			
			detail := NamespaceDetail{
				Name:        ns.Name,
				Status:      string(ns.Status.Phase),
				CreatedAt:   ns.CreationTimestamp.String(),
				Labels:      ns.Labels,
				Annotations: ns.Annotations,
			}
			namespaces = append(namespaces, detail)
		} else {
			// Get all namespaces
			nsList, err := k8sClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
			if err != nil {
				return &mcp.CallToolResultFor[GetNamespaceDetailsResult]{
					Content: []mcp.Content{&mcp.TextContent{
						Text: fmt.Sprintf("Error listing namespaces: %v", err),
					}},
				}, nil
			}
			
			for _, ns := range nsList.Items {
				detail := NamespaceDetail{
					Name:        ns.Name,
					Status:      string(ns.Status.Phase),
					CreatedAt:   ns.CreationTimestamp.String(),
					Labels:      ns.Labels,
					Annotations: ns.Annotations,
				}
				namespaces = append(namespaces, detail)
			}
		}


		// Format output
		var output string
		if len(namespaces) == 1 {
			ns := namespaces[0]
			output = fmt.Sprintf("Namespace: %s\nStatus: %s\nCreated: %s\nLabels: %v\nAnnotations: %v", 
				ns.Name, ns.Status, ns.CreatedAt, ns.Labels, ns.Annotations)
		} else {
			output = fmt.Sprintf("Found %d namespaces with details:\n", len(namespaces))
			for _, ns := range namespaces {
				output += fmt.Sprintf("\n• %s (Status: %s, Created: %s)\n  Labels: %v\n  Annotations: %v\n", 
					ns.Name, ns.Status, ns.CreatedAt, ns.Labels, ns.Annotations)
			}
		}

		return &mcp.CallToolResultFor[GetNamespaceDetailsResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: output,
			}},
		}, nil
	}
}

// registerTools registers all available MCP tools
func registerTools(server *mcp.Server, k8sClient *kubernetes.Clientset) {
	// Basic Kubernetes connectivity test
	mcp.AddTool(server, &mcp.Tool{
		Name:        "test_k8s_connection",
		Description: "Test connectivity to the Kubernetes cluster",
	}, testK8sConnection(k8sClient))

	// List namespaces tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_namespaces", 
		Description: "List all namespaces in the Kubernetes cluster",
	}, listNamespaces(k8sClient))

	// Get namespace details tool  
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_namespace_details",
		Description: "Get detailed information about namespaces (all or specific namespace)",
	}, getNamespaceDetails(k8sClient))

	log.Println("Registered MCP tools: test_k8s_connection, list_namespaces, get_namespace_details")
}