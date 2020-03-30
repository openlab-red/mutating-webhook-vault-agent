package kube

// PatchOperation defines the Kubernetes patch json strategy
type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}
