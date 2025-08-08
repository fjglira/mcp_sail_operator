package k8s

import (
	"bufio"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	"github.com/frherrer/mcp-sail-operator/pkg/types"
)

// ListPods lists pods in the cluster with optional namespace and label filtering
func ListPods(k8sClient *kubernetes.Clientset) func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[types.ListPodsParams]) (*mcp.CallToolResultFor[types.ListPodsResult], error) {
	return func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[types.ListPodsParams]) (*mcp.CallToolResultFor[types.ListPodsResult], error) {
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
			return &mcp.CallToolResultFor[types.ListPodsResult]{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error listing pods: %v", err),
				}},
			}, nil
		}

		var pods []types.PodInfo
		for _, pod := range podList.Items {
			podInfo := types.PodInfo{
				Name:      pod.Name,
				Namespace: pod.Namespace,
				Phase:     string(pod.Status.Phase),
				NodeName:  pod.Spec.NodeName,
				PodIP:     pod.Status.PodIP,
				Labels:    pod.Labels,
				CreatedAt: pod.CreationTimestamp.String(),
				Age:       formatAge(pod.CreationTimestamp.Time),
			}

			// Calculate ready containers
			readyCount := 0
			totalCount := len(pod.Status.ContainerStatuses)
			restarts := int32(0)

			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.Ready {
					readyCount++
				}
				restarts += containerStatus.RestartCount
			}

			podInfo.Ready = fmt.Sprintf("%d/%d", readyCount, totalCount)
			podInfo.Restarts = restarts

			// Determine overall status
			if pod.Status.Phase == corev1.PodRunning && readyCount == totalCount {
				podInfo.Status = "Running"
			} else if pod.Status.Phase == corev1.PodPending {
				podInfo.Status = "Pending"
			} else if pod.Status.Phase == corev1.PodFailed {
				podInfo.Status = "Failed"
			} else if pod.Status.Phase == corev1.PodSucceeded {
				podInfo.Status = "Completed"
			} else {
				podInfo.Status = string(pod.Status.Phase)
			}

			pods = append(pods, podInfo)
		}

		// Format output
		var output string
		if len(pods) == 0 {
			output = "No pods found"
			if params.Arguments.Namespace != "" {
				output += fmt.Sprintf(" in namespace '%s'", params.Arguments.Namespace)
			}
			if params.Arguments.LabelSelector != "" {
				output += fmt.Sprintf(" with label selector '%s'", params.Arguments.LabelSelector)
			}
		} else {
			output = fmt.Sprintf("Found %d pods:\n\n", len(pods))
			output += fmt.Sprintf("%-30s %-15s %-10s %-8s %-10s %-10s %s\n", 
				"NAME", "NAMESPACE", "STATUS", "READY", "RESTARTS", "AGE", "NODE")
			output += strings.Repeat("-", 100) + "\n"

			for _, pod := range pods {
				output += fmt.Sprintf("%-30s %-15s %-10s %-8s %-10d %-10s %s\n",
					truncateString(pod.Name, 29),
					pod.Namespace,
					pod.Status,
					pod.Ready,
					pod.Restarts,
					pod.Age,
					pod.NodeName,
				)
			}
		}

		return &mcp.CallToolResultFor[types.ListPodsResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: output,
			}},
		}, nil
	}
}

// ListServices lists services in the cluster with optional namespace and label filtering
func ListServices(k8sClient *kubernetes.Clientset) func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[types.ListServicesParams]) (*mcp.CallToolResultFor[types.ListServicesResult], error) {
	return func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[types.ListServicesParams]) (*mcp.CallToolResultFor[types.ListServicesResult], error) {
		listOptions := metav1.ListOptions{}
		if params.Arguments.LabelSelector != "" {
			listOptions.LabelSelector = params.Arguments.LabelSelector
		}

		var serviceList *corev1.ServiceList
		var err error

		if params.Arguments.Namespace != "" {
			serviceList, err = k8sClient.CoreV1().Services(params.Arguments.Namespace).List(ctx, listOptions)
		} else {
			serviceList, err = k8sClient.CoreV1().Services("").List(ctx, listOptions)
		}

		if err != nil {
			return &mcp.CallToolResultFor[types.ListServicesResult]{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error listing services: %v", err),
				}},
			}, nil
		}

		var services []types.ServiceInfo
		for _, svc := range serviceList.Items {
			serviceInfo := types.ServiceInfo{
				Name:      svc.Name,
				Namespace: svc.Namespace,
				Type:      string(svc.Spec.Type),
				ClusterIP: svc.Spec.ClusterIP,
				Labels:    svc.Labels,
				Selector:  svc.Spec.Selector,
				CreatedAt: svc.CreationTimestamp.String(),
			}

			// Process external IPs
			for _, ip := range svc.Status.LoadBalancer.Ingress {
				if ip.IP != "" {
					serviceInfo.ExternalIP = append(serviceInfo.ExternalIP, ip.IP)
				}
				if ip.Hostname != "" {
					serviceInfo.ExternalIP = append(serviceInfo.ExternalIP, ip.Hostname)
				}
			}

			// Process ports
			for _, port := range svc.Spec.Ports {
				servicePort := types.ServicePort{
					Name:     port.Name,
					Port:     port.Port,
					Protocol: string(port.Protocol),
					NodePort: port.NodePort,
				}

				if port.TargetPort.Type == intstr.Int {
					servicePort.TargetPort = strconv.Itoa(int(port.TargetPort.IntVal))
				} else {
					servicePort.TargetPort = port.TargetPort.StrVal
				}

				serviceInfo.Ports = append(serviceInfo.Ports, servicePort)
			}

			services = append(services, serviceInfo)
		}

		// Format output
		var output string
		if len(services) == 0 {
			output = "No services found"
			if params.Arguments.Namespace != "" {
				output += fmt.Sprintf(" in namespace '%s'", params.Arguments.Namespace)
			}
			if params.Arguments.LabelSelector != "" {
				output += fmt.Sprintf(" with label selector '%s'", params.Arguments.LabelSelector)
			}
		} else {
			output = fmt.Sprintf("Found %d services:\n\n", len(services))
			output += fmt.Sprintf("%-30s %-15s %-12s %-15s %-20s %s\n", 
				"NAME", "NAMESPACE", "TYPE", "CLUSTER-IP", "EXTERNAL-IP", "PORTS")
			output += strings.Repeat("-", 110) + "\n"

			for _, svc := range services {
				externalIP := "<none>"
				if len(svc.ExternalIP) > 0 {
					externalIP = strings.Join(svc.ExternalIP, ",")
				}

				ports := ""
				for i, port := range svc.Ports {
					if i > 0 {
						ports += ","
					}
					ports += fmt.Sprintf("%d/%s", port.Port, port.Protocol)
				}

				output += fmt.Sprintf("%-30s %-15s %-12s %-15s %-20s %s\n",
					truncateString(svc.Name, 29),
					svc.Namespace,
					svc.Type,
					svc.ClusterIP,
					truncateString(externalIP, 19),
					ports,
				)
			}
		}

		return &mcp.CallToolResultFor[types.ListServicesResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: output,
			}},
		}, nil
	}
}

// ListDeployments lists deployments in the cluster with optional namespace and label filtering
func ListDeployments(k8sClient *kubernetes.Clientset) func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[types.ListDeploymentsParams]) (*mcp.CallToolResultFor[types.ListDeploymentsResult], error) {
	return func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[types.ListDeploymentsParams]) (*mcp.CallToolResultFor[types.ListDeploymentsResult], error) {
		listOptions := metav1.ListOptions{}
		if params.Arguments.LabelSelector != "" {
			listOptions.LabelSelector = params.Arguments.LabelSelector
		}

		var deploymentList *appsv1.DeploymentList
		var err error

		if params.Arguments.Namespace != "" {
			deploymentList, err = k8sClient.AppsV1().Deployments(params.Arguments.Namespace).List(ctx, listOptions)
		} else {
			deploymentList, err = k8sClient.AppsV1().Deployments("").List(ctx, listOptions)
		}

		if err != nil {
			return &mcp.CallToolResultFor[types.ListDeploymentsResult]{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error listing deployments: %v", err),
				}},
			}, nil
		}

		var deployments []types.DeploymentInfo
		for _, deploy := range deploymentList.Items {
			deploymentInfo := types.DeploymentInfo{
				Name:      deploy.Name,
				Namespace: deploy.Namespace,
				Ready:     fmt.Sprintf("%d/%d", deploy.Status.ReadyReplicas, deploy.Status.Replicas),
				UpToDate:  deploy.Status.UpdatedReplicas,
				Available: deploy.Status.AvailableReplicas,
				Labels:    deploy.Labels,
				CreatedAt: deploy.CreationTimestamp.String(),
				Age:       formatAge(deploy.CreationTimestamp.Time),
			}

			// Get deployment strategy
			if deploy.Spec.Strategy.Type != "" {
				deploymentInfo.Strategy = string(deploy.Spec.Strategy.Type)
			}

			deployments = append(deployments, deploymentInfo)
		}

		// Format output
		var output string
		if len(deployments) == 0 {
			output = "No deployments found"
			if params.Arguments.Namespace != "" {
				output += fmt.Sprintf(" in namespace '%s'", params.Arguments.Namespace)
			}
			if params.Arguments.LabelSelector != "" {
				output += fmt.Sprintf(" with label selector '%s'", params.Arguments.LabelSelector)
			}
		} else {
			output = fmt.Sprintf("Found %d deployments:\n\n", len(deployments))
			output += fmt.Sprintf("%-30s %-15s %-8s %-10s %-10s %s\n", 
				"NAME", "NAMESPACE", "READY", "UP-TO-DATE", "AVAILABLE", "AGE")
			output += strings.Repeat("-", 90) + "\n"

			for _, deploy := range deployments {
				output += fmt.Sprintf("%-30s %-15s %-8s %-10d %-10d %s\n",
					truncateString(deploy.Name, 29),
					deploy.Namespace,
					deploy.Ready,
					deploy.UpToDate,
					deploy.Available,
					deploy.Age,
				)
			}
		}

		return &mcp.CallToolResultFor[types.ListDeploymentsResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: output,
			}},
		}, nil
	}
}

// ListConfigMaps lists configmaps in the cluster with optional namespace and label filtering
func ListConfigMaps(k8sClient *kubernetes.Clientset) func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[types.ListConfigMapsParams]) (*mcp.CallToolResultFor[types.ListConfigMapsResult], error) {
	return func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[types.ListConfigMapsParams]) (*mcp.CallToolResultFor[types.ListConfigMapsResult], error) {
		listOptions := metav1.ListOptions{}
		if params.Arguments.LabelSelector != "" {
			listOptions.LabelSelector = params.Arguments.LabelSelector
		}

		var configMapList *corev1.ConfigMapList
		var err error

		if params.Arguments.Namespace != "" {
			configMapList, err = k8sClient.CoreV1().ConfigMaps(params.Arguments.Namespace).List(ctx, listOptions)
		} else {
			configMapList, err = k8sClient.CoreV1().ConfigMaps("").List(ctx, listOptions)
		}

		if err != nil {
			return &mcp.CallToolResultFor[types.ListConfigMapsResult]{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error listing configmaps: %v", err),
				}},
			}, nil
		}

		var configMaps []types.ConfigMapInfo
		for _, cm := range configMapList.Items {
			var keys []string
			for key := range cm.Data {
				keys = append(keys, key)
			}

			configMapInfo := types.ConfigMapInfo{
				Name:      cm.Name,
				Namespace: cm.Namespace,
				DataCount: len(cm.Data),
				Labels:    cm.Labels,
				Keys:      keys,
				CreatedAt: cm.CreationTimestamp.String(),
			}

			configMaps = append(configMaps, configMapInfo)
		}

		// Format output
		var output string
		if len(configMaps) == 0 {
			output = "No configmaps found"
			if params.Arguments.Namespace != "" {
				output += fmt.Sprintf(" in namespace '%s'", params.Arguments.Namespace)
			}
			if params.Arguments.LabelSelector != "" {
				output += fmt.Sprintf(" with label selector '%s'", params.Arguments.LabelSelector)
			}
		} else {
			output = fmt.Sprintf("Found %d configmaps:\n\n", len(configMaps))
			output += fmt.Sprintf("%-30s %-15s %-5s %s\n", 
				"NAME", "NAMESPACE", "DATA", "KEYS")
			output += strings.Repeat("-", 80) + "\n"

			for _, cm := range configMaps {
				keys := strings.Join(cm.Keys, ",")
				if len(keys) > 40 {
					keys = keys[:37] + "..."
				}

				output += fmt.Sprintf("%-30s %-15s %-5d %s\n",
					truncateString(cm.Name, 29),
					cm.Namespace,
					cm.DataCount,
					keys,
				)
			}
		}

		return &mcp.CallToolResultFor[types.ListConfigMapsResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: output,
			}},
		}, nil
	}
}

// Helper functions

// formatAge formats a time duration since creation into a human-readable string
func formatAge(createdAt time.Time) string {
	duration := time.Since(createdAt)
	
	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	} else if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	} else {
		return fmt.Sprintf("%dd", int(duration.Hours()/24))
	}
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// GetPodLogs gets logs from a specific pod and container
func GetPodLogs(k8sClient *kubernetes.Clientset) func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[types.GetPodLogsParams]) (*mcp.CallToolResultFor[types.GetPodLogsResult], error) {
	return func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[types.GetPodLogsParams]) (*mcp.CallToolResultFor[types.GetPodLogsResult], error) {
		// Validate required parameters
		if params.Arguments.Namespace == "" {
			return &mcp.CallToolResultFor[types.GetPodLogsResult]{
				Content: []mcp.Content{&mcp.TextContent{
					Text: "Error: namespace parameter is required",
				}},
			}, nil
		}

		if params.Arguments.PodName == "" {
			return &mcp.CallToolResultFor[types.GetPodLogsResult]{
				Content: []mcp.Content{&mcp.TextContent{
					Text: "Error: pod_name parameter is required",
				}},
			}, nil
		}

		// Set up log options
		logOptions := &corev1.PodLogOptions{
			Follow:    params.Arguments.Follow,
			Previous:  params.Arguments.Previous,
		}

		// Set container if specified
		if params.Arguments.Container != "" {
			logOptions.Container = params.Arguments.Container
		}

		// Set tail lines if specified (default to 50 if not specified)
		if params.Arguments.Lines > 0 {
			logOptions.TailLines = &params.Arguments.Lines
		} else {
			defaultLines := int64(50)
			logOptions.TailLines = &defaultLines
		}

		// Set since seconds if specified
		if params.Arguments.SinceSeconds > 0 {
			logOptions.SinceSeconds = &params.Arguments.SinceSeconds
		}

		// Get the logs
		req := k8sClient.CoreV1().Pods(params.Arguments.Namespace).GetLogs(params.Arguments.PodName, logOptions)
		
		podLogs, err := req.Stream(ctx)
		if err != nil {
			return &mcp.CallToolResultFor[types.GetPodLogsResult]{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error getting logs for pod '%s' in namespace '%s': %v", 
						params.Arguments.PodName, params.Arguments.Namespace, err),
				}},
			}, nil
		}
		defer podLogs.Close()

		// Read the logs
		var logLines []string
		scanner := bufio.NewScanner(podLogs)
		for scanner.Scan() {
			logLines = append(logLines, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			return &mcp.CallToolResultFor[types.GetPodLogsResult]{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error reading logs for pod '%s' in namespace '%s': %v", 
						params.Arguments.PodName, params.Arguments.Namespace, err),
				}},
			}, nil
		}

		// Format output
		var output string
		if len(logLines) == 0 {
			output = fmt.Sprintf("No logs found for pod '%s' in namespace '%s'", 
				params.Arguments.PodName, params.Arguments.Namespace)
			if params.Arguments.Container != "" {
				output += fmt.Sprintf(" (container: %s)", params.Arguments.Container)
			}
		} else {
			header := fmt.Sprintf("=== Logs for pod '%s' in namespace '%s' ===", 
				params.Arguments.PodName, params.Arguments.Namespace)
			if params.Arguments.Container != "" {
				header += fmt.Sprintf(" (container: %s)", params.Arguments.Container)
			}
			header += fmt.Sprintf("\nShowing last %d lines:\n\n", len(logLines))
			
			output = header + strings.Join(logLines, "\n")
		}

		return &mcp.CallToolResultFor[types.GetPodLogsResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: output,
			}},
		}, nil
	}
}