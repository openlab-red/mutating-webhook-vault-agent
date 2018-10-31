package webhook

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/admission/v1beta1"
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"strings"
)

func Client() (*kubernetes.Clientset) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}
func Pod(raw []byte, pod *corev1.Pod) (error) {

	log.Debugf("Object: %v", string(raw))
	if err := json.Unmarshal(raw, pod); err != nil {
		log.Errorln(err)
		return err
	}
	log.Debugf("Pod: %v", pod)
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
	{
		FindVolumeMount(sidecarConfig.Containers[0].VolumeMounts, "vault-agent-volume")
	}

	log.Debugf("VolumeMounts: %v", volumeMounts)

	patch = append(patch, AddVolumeMount(pod.Spec.Containers[0].VolumeMounts, volumeMounts, "/spec/containers/0/volumeMounts")...)

	patch = append(patch, AddContainer(pod.Spec.Containers, sidecarConfig.Containers, "/spec/containers")...)
	patch = append(patch, AddVolume(pod.Spec.Volumes, sidecarConfig.Volumes, "/spec/volumes")...)
	patch = append(patch, UpdateAnnotation(pod.Annotations, annotations)...)

	log.Debugf("Patch: %v", patch)
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

func PotentialNamespace(req *v1beta1.AdmissionRequest, pod *corev1.Pod) (string) {
	if pod.ObjectMeta.Namespace == "" {
		return req.Namespace
	}
	return pod.ObjectMeta.Namespace
}

func FindTokenVolumeName(volumes []corev1.Volume) (string) {
	for _, vol := range volumes {
		if strings.Contains(vol.Name, "token") && vol.VolumeSource.Secret != nil {
			return vol.Name
		}
	}
	return ""
}

func FindVolumeMount(volumes []corev1.VolumeMount, name string) (*corev1.VolumeMount) {
	for _, vol := range volumes {
		if strings.Contains(vol.Name, name) {
			log.Debugln("VolumeMount found", vol.Name)
			return &vol
		}
	}
	return nil
}
