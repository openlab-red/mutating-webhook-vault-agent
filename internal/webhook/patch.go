package webhook

import (
	"encoding/json"

	"github.com/openlab-red/mutating-webhook-vault-agent/pkg/kube"
	corev1 "k8s.io/api/core/v1"
)

// CreatePatch to inject the change
func CreatePatch(pod *corev1.Pod, sidecarInject *SidecarInject, annotations map[string]string) ([]byte, error) {
	var patch []kube.PatchOperation

	log.Debugln("VolumeMounts:", sidecarInject.VolumeMount)
	patch = append(patch, kube.AddVolumeMount(pod.Spec.Containers[0].VolumeMounts, sidecarInject.VolumeMount, "/spec/containers/0/volumeMounts")...)
	patch = append(patch, kube.AddContainer(pod.Spec.Containers, sidecarInject.Containers, "/spec/containers")...)
	patch = append(patch, kube.AddContainer(pod.Spec.InitContainers, sidecarInject.InitContainers, "/spec/initContainers")...)
	patch = append(patch, kube.AddVolume(pod.Spec.Volumes, sidecarInject.Volumes, "/spec/volumes")...)
	patch = append(patch, kube.UpdateAnnotation(pod.Annotations, annotations)...)

	log.Debugf("Patch: %v", patch)
	return json.Marshal(patch)
}
