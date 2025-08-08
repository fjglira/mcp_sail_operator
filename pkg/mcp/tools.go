package mcp

import (
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	k8shandlers "github.com/frherrer/mcp-sail-operator/pkg/handlers/k8s"
	sailoperatorhandlers "github.com/frherrer/mcp-sail-operator/pkg/handlers/sailoperator"
)

// RegisterAllTools registers all available MCP tools with the server
func RegisterAllTools(server *mcp.Server, k8sClient *kubernetes.Clientset, dynamicClient dynamic.Interface) {
	registerK8sTools(server, k8sClient)
	registerSailOperatorTools(server, dynamicClient)

	log.Println("Registered all MCP tools")
}

// registerK8sTools registers Kubernetes-related MCP tools
func registerK8sTools(server *mcp.Server, k8sClient *kubernetes.Clientset) {
	// Basic Kubernetes connectivity test
	mcp.AddTool(server, &mcp.Tool{
		Name:        "test_k8s_connection",
		Description: "Test connectivity to the Kubernetes cluster",
	}, k8shandlers.TestConnection(k8sClient))

	// List namespaces tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_namespaces",
		Description: "List all namespaces in the Kubernetes cluster",
	}, k8shandlers.ListNamespaces(k8sClient))

	// Get namespace details tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_namespace_details",
		Description: "Get detailed information about namespaces (all or specific namespace)",
	}, k8shandlers.GetNamespaceDetails(k8sClient))

	// List pods tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_pods",
		Description: "List pods in the cluster with optional namespace and label filtering",
	}, k8shandlers.ListPods(k8sClient))

	// List services tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_services",
		Description: "List services in the cluster with optional namespace and label filtering",
	}, k8shandlers.ListServices(k8sClient))

	// List deployments tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_deployments",
		Description: "List deployments in the cluster with optional namespace and label filtering",
	}, k8shandlers.ListDeployments(k8sClient))

	// List configmaps tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_configmaps",
		Description: "List configmaps in the cluster with optional namespace and label filtering",
	}, k8shandlers.ListConfigMaps(k8sClient))

	// List events tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_events",
		Description: "List recent Kubernetes events with optional selectors",
	}, k8shandlers.ListEvents(k8sClient))

	// Get pod logs tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_pod_logs",
		Description: "Get logs from a specific pod and optionally a specific container",
	}, k8shandlers.GetPodLogs(k8sClient))

	// Check mesh workloads tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "check_mesh_workloads",
		Description: "Check the status of workloads in the Istio mesh including sidecar injection status",
	}, k8shandlers.CheckMeshWorkloads(k8sClient))

	log.Println("Registered Kubernetes tools: test_k8s_connection, list_namespaces, get_namespace_details, list_pods, list_services, list_deployments, list_configmaps, list_events, get_pod_logs, check_mesh_workloads")
}

// registerSailOperatorTools registers Sail Operator CRD-related MCP tools
func registerSailOperatorTools(server *mcp.Server, dynamicClient dynamic.Interface) {
	// List Sail Operator resources
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_sailoperator_resources",
		Description: "List Sail Operator CRD resources (Istio, IstioRevision, IstioCNI, ZTunnel)",
	}, sailoperatorhandlers.ListSailOperatorResources(dynamicClient))

	// Get Istio status
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_istio_status",
		Description: "Get detailed status information about Istio installations",
	}, sailoperatorhandlers.GetIstioStatus(dynamicClient))

	// Check Sail Operator health
	mcp.AddTool(server, &mcp.Tool{
		Name:        "check_sailoperator_health",
		Description: "Perform comprehensive health checks on Sail Operator managed resources",
	}, sailoperatorhandlers.CheckSailOperatorHealth(dynamicClient))

	log.Println("Registered Sail Operator tools: list_sailoperator_resources, get_istio_status, check_sailoperator_health")
}
