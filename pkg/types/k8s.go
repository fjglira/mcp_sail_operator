package types

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

// ListPodsParams represents parameters for listing pods
type ListPodsParams struct {
	Namespace     string `json:"namespace,omitempty"`
	LabelSelector string `json:"label_selector,omitempty"`
}

// PodInfo represents information about a pod
type PodInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Status    string            `json:"status"`
	Phase     string            `json:"phase"`
	NodeName  string            `json:"node_name,omitempty"`
	PodIP     string            `json:"pod_ip,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	Ready     string            `json:"ready"`
	Restarts  int32             `json:"restarts"`
	Age       string            `json:"age"`
	CreatedAt string            `json:"created_at"`
}

// ListPodsResult represents the result of listing pods
type ListPodsResult struct {
	Status string    `json:"status"`
	Pods   []PodInfo `json:"pods,omitempty"`
	Count  int       `json:"count,omitempty"`
	Error  string    `json:"error,omitempty"`
}

// ListServicesParams represents parameters for listing services
type ListServicesParams struct {
	Namespace     string `json:"namespace,omitempty"`
	LabelSelector string `json:"label_selector,omitempty"`
}

// ServiceInfo represents information about a service
type ServiceInfo struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Type       string            `json:"type"`
	ClusterIP  string            `json:"cluster_ip,omitempty"`
	ExternalIP []string          `json:"external_ip,omitempty"`
	Ports      []ServicePort     `json:"ports,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	Selector   map[string]string `json:"selector,omitempty"`
	CreatedAt  string            `json:"created_at"`
}

// ServicePort represents a service port
type ServicePort struct {
	Name       string `json:"name,omitempty"`
	Port       int32  `json:"port"`
	TargetPort string `json:"target_port,omitempty"`
	Protocol   string `json:"protocol"`
	NodePort   int32  `json:"node_port,omitempty"`
}

// ListServicesResult represents the result of listing services
type ListServicesResult struct {
	Status   string        `json:"status"`
	Services []ServiceInfo `json:"services,omitempty"`
	Count    int           `json:"count,omitempty"`
	Error    string        `json:"error,omitempty"`
}

// ListDeploymentsParams represents parameters for listing deployments
type ListDeploymentsParams struct {
	Namespace     string `json:"namespace,omitempty"`
	LabelSelector string `json:"label_selector,omitempty"`
}

// DeploymentInfo represents information about a deployment
type DeploymentInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Ready     string            `json:"ready"`
	UpToDate  int32             `json:"up_to_date"`
	Available int32             `json:"available"`
	Age       string            `json:"age"`
	Labels    map[string]string `json:"labels,omitempty"`
	Strategy  string            `json:"strategy,omitempty"`
	CreatedAt string            `json:"created_at"`
}

// ListDeploymentsResult represents the result of listing deployments
type ListDeploymentsResult struct {
	Status      string           `json:"status"`
	Deployments []DeploymentInfo `json:"deployments,omitempty"`
	Count       int              `json:"count,omitempty"`
	Error       string           `json:"error,omitempty"`
}

// ListConfigMapsParams represents parameters for listing configmaps
type ListConfigMapsParams struct {
	Namespace     string `json:"namespace,omitempty"`
	LabelSelector string `json:"label_selector,omitempty"`
}

// ConfigMapInfo represents information about a configmap
type ConfigMapInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	DataCount int               `json:"data_count"`
	Labels    map[string]string `json:"labels,omitempty"`
	Keys      []string          `json:"keys,omitempty"`
	CreatedAt string            `json:"created_at"`
}

// ListConfigMapsResult represents the result of listing configmaps
type ListConfigMapsResult struct {
	Status     string          `json:"status"`
	ConfigMaps []ConfigMapInfo `json:"configmaps,omitempty"`
	Count      int             `json:"count,omitempty"`
	Error      string          `json:"error,omitempty"`
}

// GetPodLogsParams represents parameters for getting pod logs
type GetPodLogsParams struct {
	Namespace    string `json:"namespace"`
	PodName      string `json:"pod_name"`
	Container    string `json:"container,omitempty"`
	Lines        int64  `json:"lines,omitempty"`
	Follow       bool   `json:"follow,omitempty"`
	Previous     bool   `json:"previous,omitempty"`
	SinceSeconds int64  `json:"since_seconds,omitempty"`
}

// GetPodLogsResult represents the result of getting pod logs
type GetPodLogsResult struct {
	Status string `json:"status"`
	Logs   string `json:"logs,omitempty"`
	Error  string `json:"error,omitempty"`
}

// CheckMeshWorkloadsParams represents parameters for checking mesh workloads
type CheckMeshWorkloadsParams struct {
	Namespace     string `json:"namespace,omitempty"`
	LabelSelector string `json:"label_selector,omitempty"`
}

// WorkloadInfo represents information about a workload in the mesh
type WorkloadInfo struct {
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	Kind            string            `json:"kind"`
	SidecarInjected bool              `json:"sidecar_injected"`
	SidecarReady    bool              `json:"sidecar_ready"`
	MeshStatus      string            `json:"mesh_status"`
	Labels          map[string]string `json:"labels,omitempty"`
	Annotations     map[string]string `json:"annotations,omitempty"`
	Issues          []string          `json:"issues,omitempty"`
}

// CheckMeshWorkloadsResult represents the result of checking mesh workloads
type CheckMeshWorkloadsResult struct {
	Status    string         `json:"status"`
	Workloads []WorkloadInfo `json:"workloads,omitempty"`
	Count     int            `json:"count,omitempty"`
	Error     string         `json:"error,omitempty"`
}

// ListEventsParams represents parameters for listing Events
type ListEventsParams struct {
	Namespace         string `json:"namespace,omitempty"`
	FieldSelector     string `json:"field_selector,omitempty"`
	InvolvedKind      string `json:"involved_kind,omitempty"`
	InvolvedName      string `json:"involved_name,omitempty"`
	InvolvedNamespace string `json:"involved_namespace,omitempty"`
	Type              string `json:"type,omitempty"` // Normal|Warning
	Reason            string `json:"reason,omitempty"`
	SinceSeconds      int64  `json:"since_seconds,omitempty"`
	Limit             int32  `json:"limit,omitempty"`
}

// EventInfo represents a summarized Kubernetes event
type EventInfo struct {
	Type              string `json:"type"`
	Reason            string `json:"reason"`
	Message           string `json:"message"`
	Count             int32  `json:"count"`
	FirstSeen         string `json:"first_seen,omitempty"`
	LastSeen          string `json:"last_seen,omitempty"`
	InvolvedKind      string `json:"involved_kind,omitempty"`
	InvolvedName      string `json:"involved_name,omitempty"`
	InvolvedNamespace string `json:"involved_namespace,omitempty"`
}

// ListEventsResult represents the result of listing events
type ListEventsResult struct {
	Status string      `json:"status"`
	Events []EventInfo `json:"events,omitempty"`
	Count  int         `json:"count,omitempty"`
	Error  string      `json:"error,omitempty"`
}
