package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	mcptools "github.com/frherrer/mcp-sail-operator/pkg/mcp"
	"github.com/frherrer/mcp-sail-operator/pkg/types"
	sailoperatorhandlers "github.com/frherrer/mcp-sail-operator/pkg/handlers/sailoperator"
)

var (
	kubeconfigPath string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "mcp-sail-operator",
		Short: "MCP server for Istio Sail Operator integration",
		Long: `MCP server that provides Claude with access to Istio Sail Operator resources 
and Kubernetes cluster information.

USAGE:
  mcp-sail-operator              # Start MCP server (default)
  mcp-sail-operator logs <pod>   # Get pod logs  
  mcp-sail-operator pods         # List pods
  mcp-sail-operator health       # Check health
  mcp-sail-operator status       # Get Istio status`,
		Run: runServer,
	}

	// Add global kubeconfig flag
	rootCmd.PersistentFlags().StringVar(&kubeconfigPath, "kubeconfig", "", 
		"Path to kubeconfig file (default: ~/.kube/config or KUBECONFIG env var)")

	// Add CLI subcommands
	rootCmd.AddCommand(createLogsCommand())
	rootCmd.AddCommand(createPodsCommand())
	rootCmd.AddCommand(createHealthCommand())
	rootCmd.AddCommand(createStatusCommand())

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

// CLI Commands

// createLogsCommand creates the logs subcommand
func createLogsCommand() *cobra.Command {
	var namespace, container string
	var lines int64
	var previous bool

	cmd := &cobra.Command{
		Use:   "logs <pod-name>",
		Short: "Get logs from a pod",
		Long: `Get logs from a specific pod with optional container specification.

EXAMPLES:
  mcp-sail-operator logs istiod-6d9fbd58b-jf7db --namespace istio-system
  mcp-sail-operator logs productpage-v1-abc123 --namespace default --container productpage
  mcp-sail-operator logs my-pod --lines 100 --previous`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			podName := args[0]
			
			// Default namespace if not specified
			if namespace == "" {
				namespace = "default"
			}

			// Initialize Kubernetes client
			k8sClient, _, err := initKubernetesClients(kubeconfigPath)
			if err != nil {
				log.Fatalf("Failed to initialize Kubernetes client: %v", err)
			}

			// Get logs using the same logic as our MCP handler
			err = getPodLogsDirectly(k8sClient, namespace, podName, container, lines, previous)
			if err != nil {
				log.Fatalf("Failed to get pod logs: %v", err)
			}
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace (default: default)")
	cmd.Flags().StringVarP(&container, "container", "c", "", "Container name (optional)")
	cmd.Flags().Int64VarP(&lines, "lines", "l", 50, "Number of lines to show")
	cmd.Flags().BoolVar(&previous, "previous", false, "Show logs from previous container instance")

	return cmd
}

// createPodsCommand creates the pods subcommand
func createPodsCommand() *cobra.Command {
	var namespace, labelSelector string

	cmd := &cobra.Command{
		Use:   "pods",
		Short: "List pods in the cluster",
		Long: `List pods with optional namespace and label filtering.

EXAMPLES:
  mcp-sail-operator pods
  mcp-sail-operator pods --namespace istio-system
  mcp-sail-operator pods --label-selector app=istiod`,
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize Kubernetes client
			k8sClient, _, err := initKubernetesClients(kubeconfigPath)
			if err != nil {
				log.Fatalf("Failed to initialize Kubernetes client: %v", err)
			}

			err = listPodsDirectly(k8sClient, namespace, labelSelector)
			if err != nil {
				log.Fatalf("Failed to list pods: %v", err)
			}
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace (empty for all namespaces)")
	cmd.Flags().StringVarP(&labelSelector, "label-selector", "l", "", "Label selector")

	return cmd
}

// createHealthCommand creates the health subcommand
func createHealthCommand() *cobra.Command {
	var namespace string

	cmd := &cobra.Command{
		Use:   "health",
		Short: "Check Sail Operator and Istio health",
		Long: `Perform comprehensive health checks on Sail Operator managed resources.

EXAMPLES:
  mcp-sail-operator health
  mcp-sail-operator health --namespace istio-system`,
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize Kubernetes clients
			_, dynamicClient, err := initKubernetesClients(kubeconfigPath)
			if err != nil {
				log.Fatalf("Failed to initialize Kubernetes clients: %v", err)
			}

			err = checkHealthDirectly(dynamicClient, namespace)
			if err != nil {
				log.Fatalf("Failed to check health: %v", err)
			}
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace (empty for all namespaces)")

	return cmd
}

// createStatusCommand creates the status subcommand
func createStatusCommand() *cobra.Command {
	var namespace string

	cmd := &cobra.Command{
		Use:   "status [istio-name]",
		Short: "Get Istio installation status",
		Long: `Get detailed status information about Istio installations.

EXAMPLES:
  mcp-sail-operator status
  mcp-sail-operator status default
  mcp-sail-operator status default --namespace sail-operator`,
		Run: func(cmd *cobra.Command, args []string) {
			istioName := "default"
			if len(args) > 0 {
				istioName = args[0]
			}

			// Istio resources are cluster-scoped, so we should leave namespace empty
			// unless explicitly specified

			// Initialize Kubernetes clients
			_, dynamicClient, err := initKubernetesClients(kubeconfigPath)
			if err != nil {
				log.Fatalf("Failed to initialize Kubernetes clients: %v", err)
			}

			err = getStatusDirectly(dynamicClient, istioName, namespace)
			if err != nil {
				log.Fatalf("Failed to get status: %v", err)
			}
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace (default: sail-operator)")

	return cmd
}

// Helper functions for direct CLI execution

// getPodLogsDirectly gets logs from a pod directly
func getPodLogsDirectly(k8sClient *kubernetes.Clientset, namespace, podName, container string, lines int64, previous bool) error {
	logOptions := &corev1.PodLogOptions{
		Previous: previous,
	}

	if container != "" {
		logOptions.Container = container
	}

	if lines > 0 {
		logOptions.TailLines = &lines
	}

	req := k8sClient.CoreV1().Pods(namespace).GetLogs(podName, logOptions)
	podLogs, err := req.Stream(context.Background())
	if err != nil {
		return fmt.Errorf("error getting logs for pod '%s' in namespace '%s': %v", podName, namespace, err)
	}
	defer podLogs.Close()

	fmt.Printf("=== Logs for pod '%s' in namespace '%s' ===\n", podName, namespace)
	if container != "" {
		fmt.Printf("Container: %s\n", container)
	}
	fmt.Printf("Lines: %d\n\n", lines)

	scanner := bufio.NewScanner(podLogs)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	return scanner.Err()
}

// listPodsDirectly lists pods directly
func listPodsDirectly(k8sClient *kubernetes.Clientset, namespace, labelSelector string) error {
	listOptions := metav1.ListOptions{}
	if labelSelector != "" {
		listOptions.LabelSelector = labelSelector
	}

	var podList *corev1.PodList
	var err error

	if namespace != "" {
		podList, err = k8sClient.CoreV1().Pods(namespace).List(context.Background(), listOptions)
	} else {
		podList, err = k8sClient.CoreV1().Pods("").List(context.Background(), listOptions)
	}

	if err != nil {
		return fmt.Errorf("error listing pods: %v", err)
	}

	if len(podList.Items) == 0 {
		fmt.Println("No pods found")
		return nil
	}

	fmt.Printf("Found %d pods:\n\n", len(podList.Items))
	fmt.Printf("%-30s %-15s %-10s %-8s %-10s %s\n", 
		"NAME", "NAMESPACE", "STATUS", "READY", "RESTARTS", "AGE")
	fmt.Println(strings.Repeat("-", 90))

	for _, pod := range podList.Items {
		readyCount := 0
		totalCount := len(pod.Status.ContainerStatuses)
		restarts := int32(0)

		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.Ready {
				readyCount++
			}
			restarts += containerStatus.RestartCount
		}

		status := string(pod.Status.Phase)
		if pod.Status.Phase == corev1.PodRunning && readyCount == totalCount {
			status = "Running"
		}

		fmt.Printf("%-30s %-15s %-10s %d/%-7d %-10d %s\n",
			truncateName(pod.Name, 29),
			pod.Namespace,
			status,
			readyCount, totalCount,
			restarts,
			formatAgeSimple(pod.CreationTimestamp),
		)
	}

	return nil
}

// checkHealthDirectly checks health directly using existing MCP handler
func checkHealthDirectly(dynamicClient dynamic.Interface, namespace string) error {
	// Create mock MCP server session and params
	ctx := context.Background()
	
	// Use the existing health check handler directly
	healthHandler := sailoperatorhandlers.CheckSailOperatorHealth(dynamicClient)
	
	// Create parameters
	params := &mcp.CallToolParamsFor[types.CheckSailOperatorHealthParams]{
		Arguments: types.CheckSailOperatorHealthParams{
			Namespace: namespace,
		},
	}
	
	// Call the handler
	result, err := healthHandler(ctx, nil, params)
	if err != nil {
		return fmt.Errorf("health check failed: %v", err)
	}
	
	// Print the result
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
			fmt.Print(textContent.Text)
		}
	}
	
	return nil
}

// getStatusDirectly gets status directly using existing MCP handler
func getStatusDirectly(dynamicClient dynamic.Interface, istioName, namespace string) error {
	// Create mock MCP server session and params
	ctx := context.Background()
	
	// Use the existing status handler directly
	statusHandler := sailoperatorhandlers.GetIstioStatus(dynamicClient)
	
	// Create parameters
	params := &mcp.CallToolParamsFor[types.GetIstioStatusParams]{
		Arguments: types.GetIstioStatusParams{
			Name:      istioName,
			Namespace: namespace,
		},
	}
	
	// Call the handler
	result, err := statusHandler(ctx, nil, params)
	if err != nil {
		return fmt.Errorf("status check failed: %v", err)
	}
	
	// Print the result
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
			fmt.Print(textContent.Text)
		}
	}
	
	return nil
}

// Helper functions
func truncateName(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatAgeSimple(timestamp metav1.Time) string {
	duration := metav1.Now().Sub(timestamp.Time)
	if duration.Hours() < 1 {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	} else if duration.Hours() < 24 {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	} else {
		return fmt.Sprintf("%dd", int(duration.Hours()/24))
	}
}

