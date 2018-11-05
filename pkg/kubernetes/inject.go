package kubernetes

import (
	"strings"
	"bytes"
	"github.com/sirupsen/logrus"
	"text/template"
	"github.com/ghodss/yaml"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"encoding/json"
)

const (
	VaultAgentConfig = "vault-agent-config"
)

func injectData(data *SidecarData, config *SidecarConfig) (*SidecarInject, error) {

	sic := SidecarInject{}

	tmpl, err := executeTemplate(config.Template, data)
	if err != nil {
		return nil, err
	}

	err = unmarshalTemplate(tmpl, &sic)
	if err != nil {
		return nil, err
	}

	// TODO: seems not working the inject to volumeMounts

	var volumeMounts []corev1.VolumeMount
	volumeMount := FindVolumeMount(sic.Containers[0].VolumeMounts, "vault-agent-volume")
	volumeMounts = append(volumeMounts, volumeMount)
	sic.VolumeMount = volumeMounts

	//

	log.Debugln("SidecarInject: ", sic)
	return &sic, nil
}

func injectRequired(ignored []string, pod *corev1.Pod) bool {
	var status, inject string
	required := false
	metadata := pod.ObjectMeta

	// skip special kubernetes system namespaces
	for _, namespace := range ignored {
		if metadata.Namespace == namespace {
			return false
		}
	}

	annotations := metadata.GetAnnotations()
	log.Debugf("Annotations: %v", annotations)

	if annotations != nil {
		status = annotations[annotationStatus.name]

		log.Debugln(status)
		if strings.ToLower(status) == "injected" {
			required = false
		} else {
			inject = annotations[annotationPolicy.name]
			log.Debugln(inject)
			switch strings.ToLower(inject) {
			default:
				required = false
			case "y", "yes", "true", "on":
				required = true
			}
		}
	}

	log.WithFields(logrus.Fields{
		"name":      metadata.Name,
		"namespace": metadata.Namespace,
		"status":    status,
		"inject":    inject,
		"required":  required,
	}).Infoln("Mutation policy")

	return required
}

func ensureConfigMap(pod corev1.Pod, wk *WebHook, sidecarData *SidecarData) (*corev1.ConfigMap, error) {
	client := Client()
	configMaps := client.CoreV1().ConfigMaps(pod.Namespace)
	name := VaultAgentConfig + "-" + sidecarData.Name
	_, err := configMaps.Get(name, metav1.GetOptions{})
	if err != nil {
		data := make(map[string]string)
		tmpl, err := executeTemplate(wk.SidecarConfig.VaultAgentConfig, sidecarData)
		if err != nil {
			return nil, err
		}

		data[VaultAgentConfig] = string(tmpl.Bytes())
		annotations := make(map[string]string)
		annotations["vault-agent.vaultproject.io"] = "generated"

		configMap := corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:        name,
				Namespace:   pod.Namespace,
				Annotations: annotations,
			},
			Data: data,
		}

		return configMaps.Create(&configMap)
	}
	return nil, nil

}

func executeTemplate(source string, data interface{}) (*bytes.Buffer, error) {
	var tmpl bytes.Buffer

	funcMap := template.FuncMap{
		"valueOrDefault": valueOrDefault,
		"toJSON":         toJSON,
	}

	temp := template.New("inject")
	t := template.Must(temp.Funcs(funcMap).Parse(source))

	if err := t.Execute(&tmpl, &data); err != nil {
		log.Errorf("Failed to execute template %v %s", err, source)
		return nil, err
	}

	return &tmpl, nil
}

func unmarshalTemplate(tmpl *bytes.Buffer, target interface{}) (error) {
	log.Debugf("Template executed, %s", string(tmpl.Bytes()))

	if err := yaml.Unmarshal(tmpl.Bytes(), &target); err != nil {
		log.Errorf("Failed to unmarshal template %v %s", err, string(tmpl.Bytes()))
		return err
	}

	return nil
}

func valueOrDefault(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func toJSON(m map[string]string) string {
	if m == nil {
		return "{}"
	}

	ba, err := json.Marshal(m)
	if err != nil {
		log.Warnf("Unable to marshal %v", m)
		return "{}"
	}

	return string(ba)
}
