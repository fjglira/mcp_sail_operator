package sailoperator

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/frherrer/mcp-sail-operator/pkg/types"
)

var (
	istioGVR = schema.GroupVersionResource{
		Group:    "sailoperator.io",
		Version:  "v1",
		Resource: "istios",
	}
)

// GetIstioStatus gets detailed status information about Istio installations
func GetIstioStatus(dynamicClient dynamic.Interface) func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[types.GetIstioStatusParams]) (*mcp.CallToolResultFor[types.GetIstioStatusResult], error) {
	return func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[types.GetIstioStatusParams]) (*mcp.CallToolResultFor[types.GetIstioStatusResult], error) {
		var istios []types.IstioStatus

		if params.Arguments.Name != "" {
			// Get specific Istio resource
			namespace := params.Arguments.Namespace
			if namespace == "" {
				namespace = "istio-system" // Default namespace
			}

			istio, err := dynamicClient.Resource(istioGVR).Namespace(namespace).Get(ctx, params.Arguments.Name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					return &mcp.CallToolResultFor[types.GetIstioStatusResult]{
						Content: []mcp.Content{&mcp.TextContent{
							Text: fmt.Sprintf("Istio resource '%s' not found in namespace '%s'", params.Arguments.Name, namespace),
						}},
					}, nil
				}
				return &mcp.CallToolResultFor[types.GetIstioStatusResult]{
					Content: []mcp.Content{&mcp.TextContent{
						Text: fmt.Sprintf("Error getting Istio resource '%s': %v", params.Arguments.Name, err),
					}},
				}, nil
			}

			status := parseIstioStatus(istio)
			istios = append(istios, status)
		} else {
			// Get all Istio resources
			var istioList *unstructured.UnstructuredList
			var err error

			if params.Arguments.Namespace != "" {
				istioList, err = dynamicClient.Resource(istioGVR).Namespace(params.Arguments.Namespace).List(ctx, metav1.ListOptions{})
			} else {
				istioList, err = dynamicClient.Resource(istioGVR).List(ctx, metav1.ListOptions{})
			}

			if err != nil {
				if errors.IsNotFound(err) {
					return &mcp.CallToolResultFor[types.GetIstioStatusResult]{
						Content: []mcp.Content{&mcp.TextContent{
							Text: "Istio CRD not found. Sail Operator may not be installed.",
						}},
					}, nil
				}
				return &mcp.CallToolResultFor[types.GetIstioStatusResult]{
					Content: []mcp.Content{&mcp.TextContent{
						Text: fmt.Sprintf("Error listing Istio resources: %v", err),
					}},
				}, nil
			}

			for _, item := range istioList.Items {
				status := parseIstioStatus(&item)
				istios = append(istios, status)
			}
		}

		// Format output
		var output string
		if len(istios) == 0 {
			output = "No Istio installations found"
			if params.Arguments.Namespace != "" {
				output += fmt.Sprintf(" in namespace '%s'", params.Arguments.Namespace)
			}
		} else if len(istios) == 1 {
			istio := istios[0]
			output = formatDetailedIstioStatus(istio)
		} else {
			output = fmt.Sprintf("Found %d Istio installations:\n\n", len(istios))
			for _, istio := range istios {
				output += formatSummaryIstioStatus(istio) + "\n"
			}
		}

		return &mcp.CallToolResultFor[types.GetIstioStatusResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: output,
			}},
		}, nil
	}
}

// parseIstioStatus extracts status information from an unstructured Istio resource
func parseIstioStatus(istio *unstructured.Unstructured) types.IstioStatus {
	status := types.IstioStatus{
		Name:      istio.GetName(),
		Namespace: istio.GetNamespace(),
		CreatedAt: istio.GetCreationTimestamp().String(),
	}

	// Extract spec information
	if spec, found, _ := unstructured.NestedMap(istio.Object, "spec"); found {
		if version, ok := spec["version"].(string); ok {
			status.Version = version
		}
		if profile, ok := spec["profile"].(string); ok {
			status.Profile = profile
		}
		if updateStrategy, found, _ := unstructured.NestedMap(istio.Object, "spec", "updateStrategy"); found {
			if strategyType, ok := updateStrategy["type"].(string); ok {
				status.UpdateStrategy = strategyType
			}
		}
	}

	// Extract status information
	if statusMap, found, _ := unstructured.NestedMap(istio.Object, "status"); found {
		if state, ok := statusMap["state"].(string); ok {
			status.State = state
		}
		if activeRev, ok := statusMap["activeRevisionName"].(string); ok {
			status.ActiveRevisionName = activeRev
		}

		// Extract revision summary
		if revisions, found, _ := unstructured.NestedMap(istio.Object, "status", "revisions"); found {
			if total, ok := revisions["total"].(int64); ok {
				status.Revisions.Total = int(total)
			}
			if ready, ok := revisions["ready"].(int64); ok {
				status.Revisions.Ready = int(ready)
			}
			if inUse, ok := revisions["inUse"].(int64); ok {
				status.Revisions.InUse = int(inUse)
			}
		}

		// Extract conditions
		if conditionsRaw, found, _ := unstructured.NestedSlice(istio.Object, "status", "conditions"); found {
			var conditions []types.ResourceCondition
			for _, condRaw := range conditionsRaw {
				if condMap, ok := condRaw.(map[string]interface{}); ok {
					condition := types.ResourceCondition{}
					if t, ok := condMap["type"].(string); ok {
						condition.Type = t
					}
					if s, ok := condMap["status"].(string); ok {
						condition.Status = s
					}
					if r, ok := condMap["reason"].(string); ok {
						condition.Reason = r
					}
					if m, ok := condMap["message"].(string); ok {
						condition.Message = m
					}
					conditions = append(conditions, condition)
				}
			}
			status.Conditions = conditions
		}
	}

	return status
}

// formatDetailedIstioStatus formats detailed status for a single Istio installation
func formatDetailedIstioStatus(istio types.IstioStatus) string {
	output := fmt.Sprintf("=== Istio: %s ===\n", istio.Name)
	output += fmt.Sprintf("Namespace: %s\n", istio.Namespace)
	output += fmt.Sprintf("Version: %s\n", istio.Version)
	output += fmt.Sprintf("State: %s\n", istio.State)
	
	if istio.Profile != "" {
		output += fmt.Sprintf("Profile: %s\n", istio.Profile)
	}
	if istio.UpdateStrategy != "" {
		output += fmt.Sprintf("Update Strategy: %s\n", istio.UpdateStrategy)
	}
	if istio.ActiveRevisionName != "" {
		output += fmt.Sprintf("Active Revision: %s\n", istio.ActiveRevisionName)
	}

	// Revision summary
	if istio.Revisions.Total > 0 {
		output += fmt.Sprintf("Revisions: %d total, %d ready, %d in use\n", 
			istio.Revisions.Total, istio.Revisions.Ready, istio.Revisions.InUse)
	}

	// Conditions
	if len(istio.Conditions) > 0 {
		output += "\nConditions:\n"
		for _, cond := range istio.Conditions {
			output += fmt.Sprintf("  • %s: %s", cond.Type, cond.Status)
			if cond.Reason != "" {
				output += fmt.Sprintf(" (%s)", cond.Reason)
			}
			if cond.Message != "" {
				output += fmt.Sprintf(" - %s", cond.Message)
			}
			output += "\n"
		}
	}

	output += fmt.Sprintf("\nCreated: %s", istio.CreatedAt)
	return output
}

// formatSummaryIstioStatus formats summary status for multiple Istio installations
func formatSummaryIstioStatus(istio types.IstioStatus) string {
	status := fmt.Sprintf("• %s (namespace: %s) - Version: %s, State: %s", 
		istio.Name, istio.Namespace, istio.Version, istio.State)
	
	// Add key condition status
	for _, cond := range istio.Conditions {
		if cond.Type == "Ready" {
			status += fmt.Sprintf(" - Ready: %s", cond.Status)
			break
		}
	}
	
	return status
}