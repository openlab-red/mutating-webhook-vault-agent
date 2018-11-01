package kubernetes

import (
	corev1 "k8s.io/api/core/v1"
	"encoding/json"
)

func addContainer(target, added []corev1.Container, basePath string) (patch []PatchOperation) {
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

func addVolume(target, added []corev1.Volume, basePath string) (patch []PatchOperation) {
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

func addVolumeMount(target, added []corev1.VolumeMount, basePath string) (patch []PatchOperation) {
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

func updateAnnotation(target map[string]string, added map[string]string) (patch []PatchOperation) {
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

func createPatch(pod *corev1.Pod, sidecarInject *SidecarInject, annotations map[string]string) ([]byte, error) {
	var patch []PatchOperation

	log.Debugln("VolumeMounts:", sidecarInject.VolumeMount)
	patch = append(patch, addVolumeMount(pod.Spec.Containers[0].VolumeMounts, sidecarInject.VolumeMount, "/spec/containers/0/volumeMounts")...)
	patch = append(patch, addContainer(pod.Spec.Containers, sidecarInject.Containers, "/spec/containers")...)
	patch = append(patch, addVolume(pod.Spec.Volumes, sidecarInject.Volumes, "/spec/volumes")...)
	patch = append(patch, updateAnnotation(pod.Annotations, annotations)...)

	log.Debugf("Patch: %v", patch)
	return json.Marshal(patch)
}
