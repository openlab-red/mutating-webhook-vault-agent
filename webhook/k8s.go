package webhook

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/admission/v1beta1"
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Pod(raw []byte, pod corev1.Pod) (error) {

	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &pod); err != nil {
		log.Errorln(err)
		return err
	}
	return nil
}

func ToAdmissionResponse(err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

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

func CreatePatch(pod *corev1.Pod, sidecarConfig *Config, annotations map[string]string) ([]byte, error) {
	var patch []PatchOperation

	patch = append(patch, AddContainer(pod.Spec.Containers, sidecarConfig.Containers, "/spec/containers")...)
	patch = append(patch, AddVolume(pod.Spec.Volumes, sidecarConfig.Volumes, "/spec/volumes")...)
	//patch = append(patch, UpdateAnnotation(pod.Annotations, annotations)...)

	return json.Marshal(patch)
}

// Deal with potential empty fields, e.g., when the pod is created by a deployment
func PotentialPodName(metadata *metav1.ObjectMeta) string {
	if metadata.Name != "" {
		return metadata.Name
	}
	if metadata.GenerateName != "" {
		return metadata.GenerateName + "***** (actual name not yet known)"
	}
	return ""
}

func PotentialPodAndNamespace(req *v1beta1.AdmissionRequest, pod *corev1.Pod) {
	pod.ObjectMeta.Name = PotentialPodName(&pod.ObjectMeta)
	req.Name = pod.ObjectMeta.Name
	if pod.ObjectMeta.Namespace == "" {
		pod.ObjectMeta.Namespace = req.Namespace
	}
}
