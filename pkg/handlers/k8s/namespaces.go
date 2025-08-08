package k8s

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/frherrer/mcp-sail-operator/pkg/types"
)

// ListNamespaces lists all namespaces in the Kubernetes cluster
func ListNamespaces(k8sClient *kubernetes.Clientset) func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[types.ListNamespacesParams]) (*mcp.CallToolResultFor[types.ListNamespacesResult], error) {
	return func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[types.ListNamespacesParams]) (*mcp.CallToolResultFor[types.ListNamespacesResult], error) {
		namespaces, err := k8sClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			return &mcp.CallToolResultFor[types.ListNamespacesResult]{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error listing namespaces: %v", err),
				}},
			}, nil
		}

		var nsNames []string
		for _, ns := range namespaces.Items {
			nsNames = append(nsNames, ns.Name)
		}

		result := types.ListNamespacesResult{
			Status:     "success",
			Namespaces: nsNames,
			Count:      len(nsNames),
		}

		return &mcp.CallToolResultFor[types.ListNamespacesResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Found %d namespaces: %v", result.Count, result.Namespaces),
			}},
		}, nil
	}
}

// GetNamespaceDetails gets detailed information about namespaces
func GetNamespaceDetails(k8sClient *kubernetes.Clientset) func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[types.GetNamespaceDetailsParams]) (*mcp.CallToolResultFor[types.GetNamespaceDetailsResult], error) {
	return func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[types.GetNamespaceDetailsParams]) (*mcp.CallToolResultFor[types.GetNamespaceDetailsResult], error) {
		var namespaces []types.NamespaceDetail
		
		if params.Arguments.Namespace != "" {
			// Get specific namespace
			ns, err := k8sClient.CoreV1().Namespaces().Get(ctx, params.Arguments.Namespace, metav1.GetOptions{})
			if err != nil {
				return &mcp.CallToolResultFor[types.GetNamespaceDetailsResult]{
					Content: []mcp.Content{&mcp.TextContent{
						Text: fmt.Sprintf("Error getting namespace %s: %v", params.Arguments.Namespace, err),
					}},
				}, nil
			}
			
			detail := types.NamespaceDetail{
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
				return &mcp.CallToolResultFor[types.GetNamespaceDetailsResult]{
					Content: []mcp.Content{&mcp.TextContent{
						Text: fmt.Sprintf("Error listing namespaces: %v", err),
					}},
				}, nil
			}
			
			for _, ns := range nsList.Items {
				detail := types.NamespaceDetail{
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

		return &mcp.CallToolResultFor[types.GetNamespaceDetailsResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: output,
			}},
		}, nil
	}
}