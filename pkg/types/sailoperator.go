package types

// ListSailOperatorResourcesParams represents parameters for listing Sail Operator resources
type ListSailOperatorResourcesParams struct {
	Namespace string `json:"namespace,omitempty"`
	Resource  string `json:"resource,omitempty"` // istio, istiorevision, istiocni, ztunnel, all
}

// SailOperatorResource represents a generic Sail Operator CRD resource
type SailOperatorResource struct {
	Kind      string                 `json:"kind"`
	Name      string                 `json:"name"`
	Namespace string                 `json:"namespace"`
	Version   string                 `json:"version,omitempty"`
	State     string                 `json:"state,omitempty"`
	Conditions []ResourceCondition   `json:"conditions,omitempty"`
	CreatedAt string                 `json:"created_at"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// ResourceCondition represents a condition in a Kubernetes resource status
type ResourceCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

// ListSailOperatorResourcesResult represents the result of listing Sail Operator resources
type ListSailOperatorResourcesResult struct {
	Status    string                  `json:"status"`
	Resources []SailOperatorResource  `json:"resources,omitempty"`
	Count     int                     `json:"count,omitempty"`
	Error     string                  `json:"error,omitempty"`
}

// GetIstioStatusParams represents parameters for getting Istio status
type GetIstioStatusParams struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// IstioStatus represents the status of an Istio installation
type IstioStatus struct {
	Name                string              `json:"name"`
	Namespace           string              `json:"namespace"`
	Version             string              `json:"version"`
	State               string              `json:"state"`
	Profile             string              `json:"profile,omitempty"`
	ActiveRevisionName  string              `json:"active_revision_name,omitempty"`
	Revisions           RevisionSummary     `json:"revisions,omitempty"`
	Conditions          []ResourceCondition `json:"conditions,omitempty"`
	UpdateStrategy      string              `json:"update_strategy,omitempty"`
	CreatedAt           string              `json:"created_at"`
}

// RevisionSummary represents summary information about Istio revisions
type RevisionSummary struct {
	Total int `json:"total"`
	Ready int `json:"ready"`
	InUse int `json:"in_use"`
}

// GetIstioStatusResult represents the result of getting Istio status
type GetIstioStatusResult struct {
	Status  string        `json:"status"`
	Istios  []IstioStatus `json:"istios,omitempty"`
	Error   string        `json:"error,omitempty"`
}

// CheckSailOperatorHealthParams represents parameters for health checking
type CheckSailOperatorHealthParams struct {
	Namespace string `json:"namespace,omitempty"`
}

// HealthCheckResult represents health check results
type HealthCheckResult struct {
	Component string              `json:"component"`
	Status    string              `json:"status"`
	Reason    string              `json:"reason,omitempty"`
	Issues    []string            `json:"issues,omitempty"`
	Conditions []ResourceCondition `json:"conditions,omitempty"`
}

// CheckSailOperatorHealthResult represents the result of health checking
type CheckSailOperatorHealthResult struct {
	Status      string              `json:"status"`
	OverallHealth string            `json:"overall_health"`
	Components  []HealthCheckResult `json:"components,omitempty"`
	Summary     string              `json:"summary,omitempty"`
	Error       string              `json:"error,omitempty"`
}