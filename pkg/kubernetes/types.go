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
	SidecarConfig *SidecarConfig
	VaultConfig   *SidecarInject
}

type SidecarConfig struct {
	Template         string `json:"template"`
	VaultAgentConfig string `json:"vault-agent-config"`
}

type SidecarData struct {
	Name          string
	Container     corev1.Container
	TokenVolume   string
	VaultSecret   string
	PropertiesExt string
	VaultRole     string
}

type SidecarInject struct {
	Containers  []corev1.Container   `yaml:"containers"`
	Volumes     []corev1.Volume      `yaml:"volumes"`
	VolumeMount []corev1.VolumeMount `yaml:"volumeMounts"`
}

type registeredAnnotation struct {
	name      string
	validator annotationValidationFunc
}

type annotationValidationFunc func(value string) error
