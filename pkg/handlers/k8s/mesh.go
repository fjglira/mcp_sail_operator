package k8s

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/frherrer/mcp-sail-operator/pkg/types"
)

// CheckMeshWorkloads checks the status of workloads in the Istio mesh
func CheckMeshWorkloads(k8sClient *kubernetes.Clientset) func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[types.CheckMeshWorkloadsParams]) (*mcp.CallToolResultFor[types.CheckMeshWorkloadsResult], error) {
	return func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[types.CheckMeshWorkloadsParams]) (*mcp.CallToolResultFor[types.CheckMeshWorkloadsResult], error) {
		listOptions := metav1.ListOptions{}
		if params.Arguments.LabelSelector != "" {
			listOptions.LabelSelector = params.Arguments.LabelSelector
		}

		var podList *corev1.PodList
		var err error

		if params.Arguments.Namespace != "" {
			podList, err = k8sClient.CoreV1().Pods(params.Arguments.Namespace).List(ctx, listOptions)
		} else {
			podList, err = k8sClient.CoreV1().Pods("").List(ctx, listOptions)
		}

		if err != nil {
			return &mcp.CallToolResultFor[types.CheckMeshWorkloadsResult]{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error listing pods: %v", err),
				}},
			}, nil
		}

		var workloads []types.WorkloadInfo
		injectedCount := 0
		totalCount := 0

		for _, pod := range podList.Items {
			// Skip system pods
			if isSystemPod(&pod) {
				continue
			}

			totalCount++
			workload := analyzePodMeshStatus(&pod)
			if workload.SidecarInjected {
				injectedCount++
			}
			workloads = append(workloads, workload)
		}

		// Format output
		var output string
		if len(workloads) == 0 {
			output = "No workloads found"
			if params.Arguments.Namespace != "" {
				output += fmt.Sprintf(" in namespace '%s'", params.Arguments.Namespace)
			}
		} else {
			output = fmt.Sprintf("=== Mesh Workloads Analysis ===\n\n")
			output += fmt.Sprintf("Found %d workloads (%d with sidecars, %d without)\n\n", 
				totalCount, injectedCount, totalCount-injectedCount)
			
			output += fmt.Sprintf("%-30s %-15s %-12s %-8s %-10s %s\n", 
				"NAME", "NAMESPACE", "MESH STATUS", "SIDECAR", "READY", "ISSUES")
			output += strings.Repeat("-", 100) + "\n"

			for _, workload := range workloads {
				sidecarStatus := "❌"
				if workload.SidecarInjected {
					if workload.SidecarReady {
						sidecarStatus = "✅"
					} else {
						sidecarStatus = "⚠️"
					}
				}

				readyStatus := "❌"
				if workload.SidecarReady {
					readyStatus = "✅"
				}

				issues := "None"
				if len(workload.Issues) > 0 {
					issues = fmt.Sprintf("%d issues", len(workload.Issues))
				}

				output += fmt.Sprintf("%-30s %-15s %-12s %-8s %-10s %s\n",
					truncateString(workload.Name, 29),
					workload.Namespace,
					workload.MeshStatus,
					sidecarStatus,
					readyStatus,
					issues,
				)
			}

			// Add detailed issues
			issueCount := 0
			for _, workload := range workloads {
				if len(workload.Issues) > 0 {
					if issueCount == 0 {
						output += "\n=== Issues Found ===\n"
					}
					issueCount++
					output += fmt.Sprintf("\n%s (%s):\n", workload.Name, workload.Namespace)
					for _, issue := range workload.Issues {
						output += fmt.Sprintf("  • %s\n", issue)
					}
				}
			}

			if issueCount == 0 {
				output += "\n✅ No issues found with mesh workloads"
			}
		}

		return &mcp.CallToolResultFor[types.CheckMeshWorkloadsResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: output,
			}},
		}, nil
	}
}

// isSystemPod checks if a pod is a system pod that should be excluded from mesh analysis
func isSystemPod(pod *corev1.Pod) bool {
	systemNamespaces := map[string]bool{
		"kube-system":         true,
		"kube-public":         true,
		"kube-node-lease":     true,
		"local-path-storage":  true,
		"istio-system":        true,
		"istio-cni":          true,
		"sail-operator":      true,
	}
	
	return systemNamespaces[pod.Namespace]
}

// analyzePodMeshStatus analyzes a pod's mesh injection status
func analyzePodMeshStatus(pod *corev1.Pod) types.WorkloadInfo {
	workload := types.WorkloadInfo{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Kind:      "Pod",
		Labels:    pod.Labels,
		Annotations: pod.Annotations,
		Issues:    []string{},
	}

	// Check for sidecar injection
	sidecarInjected := false
	sidecarReady := false
	
	// Look for istio-proxy container
	for _, container := range pod.Spec.Containers {
		if container.Name == "istio-proxy" {
			sidecarInjected = true
			break
		}
	}

	// Check container status
	if sidecarInjected {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.Name == "istio-proxy" {
				sidecarReady = containerStatus.Ready
				if !containerStatus.Ready {
					workload.Issues = append(workload.Issues, 
						"Istio sidecar not ready")
				}
				break
			}
		}
	}

	// Check injection annotations
	if pod.Annotations != nil {
		injectionAnnotation := pod.Annotations["sidecar.istio.io/inject"]
		if injectionAnnotation == "false" && sidecarInjected {
			workload.Issues = append(workload.Issues, 
				"Pod has sidecar despite injection disabled")
		} else if injectionAnnotation == "true" && !sidecarInjected {
			workload.Issues = append(workload.Issues, 
				"Pod missing sidecar despite injection enabled")
		}

		// Check for istio status annotation
		if statusAnnotation, exists := pod.Annotations["sidecar.istio.io/status"]; exists && sidecarInjected {
			if !strings.Contains(statusAnnotation, "istio-proxy") {
				workload.Issues = append(workload.Issues, 
					"Istio status annotation missing proxy information")
			}
		} else if sidecarInjected && statusAnnotation == "" {
			workload.Issues = append(workload.Issues, 
				"Missing istio status annotation")
		}
	}

	// Determine mesh status
	if sidecarInjected && sidecarReady {
		workload.MeshStatus = "In Mesh"
	} else if sidecarInjected && !sidecarReady {
		workload.MeshStatus = "Mesh Issues"
	} else {
		workload.MeshStatus = "Not in Mesh"
	}

	workload.SidecarInjected = sidecarInjected
	workload.SidecarReady = sidecarReady

	return workload
}

