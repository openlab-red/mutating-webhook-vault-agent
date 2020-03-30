package webhook

import (
	corev1 "k8s.io/api/core/v1"
)

// WebHook defines the webhook configuration
type WebHook struct {
	SidecarConfig *SidecarConfig
	VaultConfig   *SidecarInject
}

// SidecarConfig defines the sidecar ConfigMap configuration
type SidecarConfig struct {
	Template           string `json:"template"`
	VaultAgentConfig   string `json:"agent.config"`
	VaultAgentTemplate string `json:"template.ctmpl"`
}

// SidecarData defines data to be injected in the template
type SidecarData struct {
	Name          string
	Container     corev1.Container
	TokenVolume   string
	VaultSecret   string
	VaultFileName string
	VaultRole     string
	VaultInit     bool
}

// SidecarInject defines the content to be injected
type SidecarInject struct {
	InitContainers []corev1.Container   `yaml:"initContainers"`
	Containers     []corev1.Container   `yaml:"containers"`
	Volumes        []corev1.Volume      `yaml:"volumes"`
	VolumeMount    []corev1.VolumeMount `yaml:"volumeMounts"`
}

type registeredAnnotation struct {
	name      string
	validator annotationValidationFunc
}

type annotationValidationFunc func(value string) error
