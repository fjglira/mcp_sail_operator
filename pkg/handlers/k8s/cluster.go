package k8s

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"k8s.io/client-go/kubernetes"

	"github.com/frherrer/mcp-sail-operator/pkg/types"
)

// TestConnection tests connectivity to the Kubernetes cluster
func TestConnection(k8sClient *kubernetes.Clientset) func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[types.TestConnectionParams]) (*mcp.CallToolResultFor[types.TestConnectionResult], error) {
	return func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[types.TestConnectionParams]) (*mcp.CallToolResultFor[types.TestConnectionResult], error) {
		// Try to get cluster version
		version, err := k8sClient.Discovery().ServerVersion()
		if err != nil {
			return &mcp.CallToolResultFor[types.TestConnectionResult]{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error connecting to Kubernetes: %v", err),
				}},
			}, nil
		}

		result := types.TestConnectionResult{
			Status:            "connected",
			KubernetesVersion: version.String(),
			ServerVersion:     version.GitVersion,
		}

		return &mcp.CallToolResultFor[types.TestConnectionResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Successfully connected to Kubernetes cluster.\nVersion: %s\nServer: %s", 
					result.KubernetesVersion, result.ServerVersion),
			}},
		}, nil
	}
}