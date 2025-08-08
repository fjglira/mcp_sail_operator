package sailoperator

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/frherrer/mcp-sail-operator/pkg/types"
)

// ListSailOperatorResources lists Sail Operator CRD resources
func ListSailOperatorResources(dynamicClient dynamic.Interface) func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[types.ListSailOperatorResourcesParams]) (*mcp.CallToolResultFor[types.ListSailOperatorResourcesResult], error) {
	return func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[types.ListSailOperatorResourcesParams]) (*mcp.CallToolResultFor[types.ListSailOperatorResourcesResult], error) {
		var resources []types.SailOperatorResource
		var totalCount int

		// Define Sail Operator CRD resource types
		crdTypes := map[string]schema.GroupVersionResource{
			"istio": {
				Group:    "sailoperator.io",
				Version:  "v1",
				Resource: "istios",
			},
			"istiorevision": {
				Group:    "sailoperator.io",
				Version:  "v1",
				Resource: "istiorevisions",
			},
			"istiocni": {
				Group:    "sailoperator.io",
				Version:  "v1",
				Resource: "istiocnis",
			},
			"ztunnel": {
				Group:    "sailoperator.io",
				Version:  "v1alpha1",
				Resource: "ztunnels",
			},
		}

		// Determine which resources to query
		resourcesToQuery := make(map[string]schema.GroupVersionResource)
		if params.Arguments.Resource == "" || params.Arguments.Resource == "all" {
			resourcesToQuery = crdTypes
		} else {
			if gvr, exists := crdTypes[strings.ToLower(params.Arguments.Resource)]; exists {
				resourcesToQuery[params.Arguments.Resource] = gvr
			} else {
				return &mcp.CallToolResultFor[types.ListSailOperatorResourcesResult]{
					Content: []mcp.Content{&mcp.TextContent{
						Text: fmt.Sprintf("Unknown resource type: %s. Available types: istio, istiorevision, istiocni, ztunnel", params.Arguments.Resource),
					}},
				}, nil
			}
		}

		// Query each resource type
		for resourceType, gvr := range resourcesToQuery {
			var resourceList *unstructured.UnstructuredList
			var err error

			if params.Arguments.Namespace != "" {
				resourceList, err = dynamicClient.Resource(gvr).Namespace(params.Arguments.Namespace).List(ctx, metav1.ListOptions{})
			} else {
				resourceList, err = dynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
			}

			if err != nil {
				if errors.IsNotFound(err) {
					// CRD might not be installed, continue with other resources
					continue
				}
				return &mcp.CallToolResultFor[types.ListSailOperatorResourcesResult]{
					Content: []mcp.Content{&mcp.TextContent{
						Text: fmt.Sprintf("Error listing %s resources: %v", resourceType, err),
					}},
				}, nil
			}

			// Process each resource
			for _, item := range resourceList.Items {
				resource := types.SailOperatorResource{
					Kind:      item.GetKind(),
					Name:      item.GetName(),
					Namespace: item.GetNamespace(),
					CreatedAt: item.GetCreationTimestamp().String(),
				}

				// Extract version from spec
				if spec, found, _ := unstructured.NestedMap(item.Object, "spec"); found {
					if version, ok := spec["version"].(string); ok {
						resource.Version = version
					}
				}

				// Extract state and conditions from status
				if status, found, _ := unstructured.NestedMap(item.Object, "status"); found {
					if state, ok := status["state"].(string); ok {
						resource.State = state
					}

					// Extract conditions
					if conditionsRaw, found, _ := unstructured.NestedSlice(item.Object, "status", "conditions"); found {
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
						resource.Conditions = conditions
					}
				}

				resources = append(resources, resource)
				totalCount++
			}
		}

		// Format output
		var output string
		if totalCount == 0 {
			output = "No Sail Operator resources found"
			if params.Arguments.Namespace != "" {
				output += fmt.Sprintf(" in namespace '%s'", params.Arguments.Namespace)
			}
			if params.Arguments.Resource != "" && params.Arguments.Resource != "all" {
				output += fmt.Sprintf(" of type '%s'", params.Arguments.Resource)
			}
		} else {
			output = fmt.Sprintf("Found %d Sail Operator resources:\n\n", totalCount)
			
			// Group by resource type
			resourcesByType := make(map[string][]types.SailOperatorResource)
			for _, res := range resources {
				resourcesByType[res.Kind] = append(resourcesByType[res.Kind], res)
			}

			for kind, resList := range resourcesByType {
				output += fmt.Sprintf("=== %s ===\n", kind)
				for _, res := range resList {
					output += fmt.Sprintf("• %s", res.Name)
					if res.Namespace != "" {
						output += fmt.Sprintf(" (namespace: %s)", res.Namespace)
					}
					if res.Version != "" {
						output += fmt.Sprintf(" - Version: %s", res.Version)
					}
					if res.State != "" {
						output += fmt.Sprintf(" - State: %s", res.State)
					}

					// Show critical conditions
					readyCondition := ""
					reconciledCondition := ""
					for _, cond := range res.Conditions {
						if cond.Type == "Ready" {
							readyCondition = fmt.Sprintf(" - Ready: %s", cond.Status)
							if cond.Reason != "" {
								readyCondition += fmt.Sprintf(" (%s)", cond.Reason)
							}
						}
						if cond.Type == "Reconciled" {
							reconciledCondition = fmt.Sprintf(" - Reconciled: %s", cond.Status)
							if cond.Reason != "" {
								reconciledCondition += fmt.Sprintf(" (%s)", cond.Reason)
							}
						}
					}
					output += readyCondition + reconciledCondition
					output += "\n"
				}
				output += "\n"
			}
		}

		return &mcp.CallToolResultFor[types.ListSailOperatorResourcesResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: output,
			}},
		}, nil
	}
}