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

// CheckSailOperatorHealth performs comprehensive health checks on Sail Operator managed resources
func CheckSailOperatorHealth(dynamicClient dynamic.Interface) func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[types.CheckSailOperatorHealthParams]) (*mcp.CallToolResultFor[types.CheckSailOperatorHealthResult], error) {
	return func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[types.CheckSailOperatorHealthParams]) (*mcp.CallToolResultFor[types.CheckSailOperatorHealthResult], error) {
		var components []types.HealthCheckResult
		var overallHealth = "Healthy"
		var healthyCount, totalCount int

		// Define components to check
		componentChecks := map[string]schema.GroupVersionResource{
			"Istio": {
				Group:    "sailoperator.io",
				Version:  "v1",
				Resource: "istios",
			},
			"IstioRevision": {
				Group:    "sailoperator.io",
				Version:  "v1",
				Resource: "istiorevisions",
			},
			"IstioCNI": {
				Group:    "sailoperator.io",
				Version:  "v1",
				Resource: "istiocnis",
			},
			"ZTunnel": {
				Group:    "sailoperator.io",
				Version:  "v1alpha1",
				Resource: "ztunnels",
			},
		}

		// Check each component type
		for componentName, gvr := range componentChecks {
			healthResult := checkComponentHealth(ctx, dynamicClient, componentName, gvr, params.Arguments.Namespace)
			components = append(components, healthResult)
			totalCount++

			if healthResult.Status == "Healthy" {
				healthyCount++
			}
		}

		// Determine overall health
		if healthyCount == 0 {
			overallHealth = "Unhealthy"
		} else if healthyCount < totalCount {
			overallHealth = "Degraded"
		}

		// Generate summary
		summary := generateHealthSummary(components, overallHealth, healthyCount, totalCount)

		// Format output
		output := fmt.Sprintf("=== Sail Operator Health Check ===\n\n")
		output += fmt.Sprintf("Overall Health: %s (%d/%d components healthy)\n\n", overallHealth, healthyCount, totalCount)

		for _, component := range components {
			output += formatComponentHealth(component)
		}

		output += fmt.Sprintf("\n%s", summary)

		return &mcp.CallToolResultFor[types.CheckSailOperatorHealthResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: output,
			}},
		}, nil
	}
}

// checkComponentHealth checks the health of a specific component type
func checkComponentHealth(ctx context.Context, dynamicClient dynamic.Interface, componentName string, gvr schema.GroupVersionResource, namespace string) types.HealthCheckResult {
	result := types.HealthCheckResult{
		Component: componentName,
		Status:    "NotFound",
		Issues:    []string{},
	}

	// List resources
	var resourceList *unstructured.UnstructuredList
	var err error

	if namespace != "" {
		resourceList, err = dynamicClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
	} else {
		resourceList, err = dynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		if errors.IsNotFound(err) {
			result.Reason = "CRD not installed"
			result.Issues = append(result.Issues, fmt.Sprintf("%s CRD is not installed or not accessible", componentName))
			return result
		}
		result.Status = "Error"
		result.Reason = "Query failed"
		result.Issues = append(result.Issues, fmt.Sprintf("Failed to query %s resources: %v", componentName, err))
		return result
	}

	if len(resourceList.Items) == 0 {
		result.Status = "NotInstalled"
		result.Reason = "No resources found"
		result.Issues = append(result.Issues, fmt.Sprintf("No %s resources are installed", componentName))
		return result
	}

	// Analyze health of found resources
	var healthyResources, totalResources int
	var resourceIssues []string
	var allConditions []types.ResourceCondition

	for _, item := range resourceList.Items {
		totalResources++
		isHealthy, issues, conditions := analyzeResourceHealth(&item)
		
		if isHealthy {
			healthyResources++
		}
		
		resourceIssues = append(resourceIssues, issues...)
		allConditions = append(allConditions, conditions...)
	}

	result.Conditions = allConditions

	// Determine overall component status
	if healthyResources == totalResources {
		result.Status = "Healthy"
		result.Reason = fmt.Sprintf("All %d resources healthy", totalResources)
	} else if healthyResources > 0 {
		result.Status = "Degraded"
		result.Reason = fmt.Sprintf("%d/%d resources healthy", healthyResources, totalResources)
		result.Issues = resourceIssues
	} else {
		result.Status = "Unhealthy"
		result.Reason = fmt.Sprintf("0/%d resources healthy", totalResources)
		result.Issues = resourceIssues
	}

	return result
}

// analyzeResourceHealth analyzes the health of a single resource
func analyzeResourceHealth(resource *unstructured.Unstructured) (bool, []string, []types.ResourceCondition) {
	var issues []string
	var conditions []types.ResourceCondition
	isHealthy := true

	resourceName := resource.GetName()
	resourceNamespace := resource.GetNamespace()
	resourceId := resourceName
	if resourceNamespace != "" {
		resourceId = fmt.Sprintf("%s/%s", resourceNamespace, resourceName)
	}

	// Check status conditions
	if conditionsRaw, found, _ := unstructured.NestedSlice(resource.Object, "status", "conditions"); found {
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

				// Check critical conditions
				if condition.Type == "Ready" && condition.Status != "True" {
					isHealthy = false
					issue := fmt.Sprintf("%s is not ready", resourceId)
					if condition.Reason != "" {
						issue += fmt.Sprintf(" (%s)", condition.Reason)
					}
					issues = append(issues, issue)
				}

				if condition.Type == "Reconciled" && condition.Status != "True" {
					isHealthy = false
					issue := fmt.Sprintf("%s reconciliation failed", resourceId)
					if condition.Reason != "" {
						issue += fmt.Sprintf(" (%s)", condition.Reason)
					}
					issues = append(issues, issue)
				}

				if condition.Type == "DependenciesHealthy" && condition.Status != "True" {
					isHealthy = false
					issue := fmt.Sprintf("%s has unhealthy dependencies", resourceId)
					if condition.Reason != "" {
						issue += fmt.Sprintf(" (%s)", condition.Reason)
					}
					issues = append(issues, issue)
				}
			}
		}
	}

	// Check state field
	if status, found, _ := unstructured.NestedMap(resource.Object, "status"); found {
		if state, ok := status["state"].(string); ok {
			if state != "Healthy" {
				isHealthy = false
				issues = append(issues, fmt.Sprintf("%s state is %s", resourceId, state))
			}
		}
	}

	return isHealthy, issues, conditions
}

// generateHealthSummary generates a summary of the health check results
func generateHealthSummary(components []types.HealthCheckResult, overallHealth string, healthyCount, totalCount int) string {
	summary := "=== Summary ===\n"

	if overallHealth == "Healthy" {
		summary += "✅ All Sail Operator components are healthy and functioning properly.\n"
	} else if overallHealth == "Degraded" {
		summary += "⚠️  Some Sail Operator components have issues that need attention.\n"
		summary += "\nComponents with issues:\n"
		for _, comp := range components {
			if comp.Status != "Healthy" {
				summary += fmt.Sprintf("  • %s: %s", comp.Component, comp.Reason)
				if len(comp.Issues) > 0 {
					summary += fmt.Sprintf(" - %s", strings.Join(comp.Issues[:1], ""))
				}
				summary += "\n"
			}
		}
	} else {
		summary += "❌ Sail Operator components are experiencing significant issues.\n"
		summary += "\nCritical issues found:\n"
		for _, comp := range components {
			if comp.Status == "Unhealthy" || comp.Status == "Error" {
				summary += fmt.Sprintf("  • %s: %s\n", comp.Component, comp.Reason)
				for _, issue := range comp.Issues {
					summary += fmt.Sprintf("    - %s\n", issue)
				}
			}
		}
	}

	return summary
}

// formatComponentHealth formats the health information for a single component
func formatComponentHealth(component types.HealthCheckResult) string {
	var statusEmoji string
	switch component.Status {
	case "Healthy":
		statusEmoji = "✅"
	case "Degraded":
		statusEmoji = "⚠️"
	case "Unhealthy", "Error":
		statusEmoji = "❌"
	case "NotFound", "NotInstalled":
		statusEmoji = "⭕"
	default:
		statusEmoji = "❓"
	}

	output := fmt.Sprintf("%s %s: %s", statusEmoji, component.Component, component.Status)
	if component.Reason != "" {
		output += fmt.Sprintf(" - %s", component.Reason)
	}
	output += "\n"

	// Add issues if any
	for _, issue := range component.Issues {
		output += fmt.Sprintf("   🔸 %s\n", issue)
	}

	return output
}