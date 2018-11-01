package kubernetes

import (
	corev1 "k8s.io/api/core/v1"
	"encoding/json"
)

func AddContainer(target, added []corev1.Container, basePath string) (patch []PatchOperation) {
	first := len(target) == 0
	var value interface{}
	for _, add := range added {
		value = add
		path := basePath
		if first {
			first = false
			value = []corev1.Container{add}
		} else {
			path = path + "/-"
		}
		patch = append(patch, PatchOperation{
			Op:    "add",
			Path:  path,
			Value: value,
		})
	}
	return patch
}

func AddVolume(target, added []corev1.Volume, basePath string) (patch []PatchOperation) {
	first := len(target) == 0
	var value interface{}
	for _, add := range added {
		value = add
		path := basePath
		if first {
			first = false
			value = []corev1.Volume{add}
		} else {
			path = path + "/-"
		}
		patch = append(patch, PatchOperation{
			Op:    "add",
			Path:  path,
			Value: value,
		})
	}
	return patch
}

func AddVolumeMount(target, added []corev1.VolumeMount, basePath string) (patch []PatchOperation) {
	first := len(target) == 0
	var value interface{}
	for _, add := range added {
		value = add
		path := basePath
		if first {
			first = false
			value = []corev1.VolumeMount{add}
		} else {
			path = path + "/-"
		}
		patch = append(patch, PatchOperation{
			Op:    "add",
			Path:  path,
			Value: value,
		})
	}

	log.Debugf("VolumeMount Patch: %v", patch)

	return patch
}

func UpdateAnnotation(target map[string]string, added map[string]string) (patch []PatchOperation) {
	for key, value := range added {
		if target == nil || target[key] == "" {
			target = map[string]string{}
			patch = append(patch, PatchOperation{
				Op:   "add",
				Path: "/metadata/annotations",
				Value: map[string]string{
					key: value,
				},
			})
		} else {
			patch = append(patch, PatchOperation{
				Op:    "replace",
				Path:  "/metadata/annotations/" + key,
				Value: value,
			})
		}
	}
	return patch
}

func CreatePatch(pod *corev1.Pod, sidecarConfig *VaultConfig, annotations map[string]string) ([]byte, error) {
	var patch []PatchOperation
	var volumeMounts []corev1.VolumeMount

	volumeMount := FindVolumeMount(sidecarConfig.Containers[0].VolumeMounts, "vault-agent-volume")
	volumeMounts = append(volumeMounts, volumeMount)
	log.Debugf("VolumeMounts: %v", volumeMounts)

	patch = append(patch, AddVolumeMount(pod.Spec.Containers[0].VolumeMounts, volumeMounts, "/spec/containers/0/volumeMounts")...)
	patch = append(patch, AddContainer(pod.Spec.Containers, sidecarConfig.Containers, "/spec/containers")...)
	patch = append(patch, AddVolume(pod.Spec.Volumes, sidecarConfig.Volumes, "/spec/volumes")...)
	patch = append(patch, UpdateAnnotation(pod.Annotations, annotations)...)

	log.Debugf("Patch: %v", patch)
	return json.Marshal(patch)
}
