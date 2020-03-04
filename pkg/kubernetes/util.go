package kubernetes

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/ghodss/yaml"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Load(file string, c interface{}) {

	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalln(err)
	}

	if err := yaml.Unmarshal(data, c); err != nil {
		log.Warnf("Failed to parse %s", string(data))
	}

	log.Debugf("New configuration: sha256sum %x", sha256.Sum256(data))
	log.Infof("SidecarConfig: %v", c)
}

func Pod(raw []byte, pod *corev1.Pod) error {

	log.Debugf("Object: %v", string(raw))
	if err := json.Unmarshal(raw, pod); err != nil {
		log.Errorln(err)
		return err
	}
	log.Debugf("Pod: %v", pod)
	return nil
}

func GetAnnotationValue(pod corev1.Pod, name *registeredAnnotation, defaultValue string) string {
	metadata := pod.ObjectMeta
	annotations := metadata.GetAnnotations()
	return name.getValueOrDefault(annotations, defaultValue)
}

func GetDeploymentName(name string) (string, error) {
	re := regexp.MustCompile("-[0-9]+")
	index := re.FindIndex([]byte(name))
	if len(index) > 0 {
		return name[:index[0]], nil
	}
	return "", errors.New(fmt.Sprintf("Wrong string format %s, expected version number", name))
}

func ToAdmissionResponseError(err error) *v1.AdmissionResponse {
	log.Errorln(err)
	return &v1.AdmissionResponse{
		Allowed: false,
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
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

func PotentialNamespace(req *v1.AdmissionRequest, pod *corev1.Pod) string {
	if pod.ObjectMeta.Namespace == "" {
		return req.Namespace
	}
	return pod.ObjectMeta.Namespace
}

func FindTokenVolumeName(volumes []corev1.Volume) string {
	for _, vol := range volumes {
		if strings.Contains(vol.Name, "token") && vol.VolumeSource.Secret != nil {
			return vol.Name
		}
	}
	return ""
}

func FindVolumeMount(volumes []corev1.VolumeMount, name string) corev1.VolumeMount {
	for _, vol := range volumes {
		if strings.Contains(vol.Name, name) {
			log.Debugln("VolumeMount found", vol.Name)
			return vol
		}
	}
	return corev1.VolumeMount{}
}

func (v *registeredAnnotation) getValueOrDefault(annotations map[string]string, defaultValue string) string {
	if val, ok := annotations[v.name]; ok {
		return val
	}
	return defaultValue
}
