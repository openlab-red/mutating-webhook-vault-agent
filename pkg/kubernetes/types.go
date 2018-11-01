package kubernetes

import (
	corev1 "k8s.io/api/core/v1"
)

type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

type WebHook struct {
	Config      *Config
	VaultConfig *VaultConfig
}

type Config struct {
	Template         string `json:"template"`
	VaultAgentConfig string `json:"vault-agent-config"`
}

type SidecarData struct {
	Container   corev1.Container
	TokenVolume string
}

type VaultConfig struct {
	Containers []corev1.Container `yaml:"containers"`
	Volumes    []corev1.Volume    `yaml:"volumes"`
}

type registeredAnnotation struct {
	name      string
	validator annotationValidationFunc
}

type annotationValidationFunc func(value string) error

type SideCarFile struct {
	Template         string `json:"template"`
	VaultAgentConfig string `json:"vault-agent-config"`
	VaultVolumeMount string `json:"vault-volume"`
}

type SideCarInject struct {
	Containers  []corev1.Container `yaml:"containers"`
	Volumes     []corev1.Volume    `yaml:"volumes"`
	VolumeMount []corev1.VolumeMount `yaml:"volumesMounts"`
}

