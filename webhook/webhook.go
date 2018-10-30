package webhook

import (
	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/api/admission/v1beta1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/kubernetes/pkg/apis/core/v1"
	"net/http"
	"github.com/sirupsen/logrus"
	"strings"
	"text/template"
	"bytes"
	"github.com/ghodss/yaml"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	alwaysValidFunc = func(value string) error {
		return nil
	}

	annotationRegistry = []*registeredAnnotation{
		{"sidecar.agent.vaultproject.io/inject", alwaysValidFunc},
		{"sidecar.agent.vaultproject.io/status", alwaysValidFunc},
	}

	annotationPolicy = annotationRegistry[0]
	annotationStatus = annotationRegistry[1]

	ignoredNamespaces = []string{
		metav1.NamespaceSystem,
		metav1.NamespacePublic,
	}
)

func (wk *WebHook) mutate(context *gin.Context) {

	ar := v1beta1.AdmissionReview{}

	if err := context.ShouldBindJSON(&ar); err == nil {
		admissionResponse := wk.admit(ar)
		admissionReview := v1beta1.AdmissionReview{}
		if admissionResponse != nil {
			admissionReview.Response = admissionResponse
			if ar.Request != nil {
				admissionReview.Response.UID = ar.Request.UID
			}
		}
		context.JSON(http.StatusOK, admissionReview)
	} else {
		context.AbortWithStatusJSON(http.StatusBadRequest, ToAdmissionResponse(err))
	}

}

func (wk *WebHook) admit(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	pod := corev1.Pod{}
	var err error

	if err = Pod(req.Object.Raw, &pod); err != nil {
		return ToAdmissionResponse(err)
	}

	pod.Name = PotentialPodName(&pod.ObjectMeta)
	pod.Namespace = PotentialNamespace(req, &pod)

	log.WithFields(logrus.Fields{
		"Kind":           req.Kind,
		"Namespace":      req.Namespace,
		"Name":           pod.Name,
		"UID":            req.UID,
		"PatchOperation": req.Operation,
		"UserInfo":       req.UserInfo,
	}).Infoln("AdmissionReview for")

	if !injectionRequired(ignoredNamespaces, &pod) {
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	wk.sidecarConfig, err = injectData(&pod, wk.config)
	if err != nil {
		return ToAdmissionResponse(err)
	}

	annotations := map[string]string{annotationStatus.name: "injected"}
	patches, err := CreatePatch(&pod, wk.sidecarConfig, annotations)

	if err != nil {
		return ToAdmissionResponse(err)
	}

	log.Debugf("AdmissionResponse: patch=%v\n", string(patches))

	return &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patches,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

func injectData(pod *corev1.Pod, config *Config) (*SideCarConfig, error) {
	var tmpl bytes.Buffer

	temp := template.New("inject")
	t, err := temp.Parse(config.Template)
	if err != nil {
		t.Execute(&tmpl, GetSecurityContext(pod.Spec.Containers[0]))
	}

	var sic SideCarConfig
	if err := yaml.Unmarshal(tmpl.Bytes(), &sic); err != nil {
		log.Warnf("Failed to unmarshall template %v %s", err, string(tmpl.Bytes()))
		return nil, err
	}
	log.Debugln("SideCarConfig: ", sic)
	return &sic, nil
}

func injectionRequired(ignored []string, pod *corev1.Pod) bool {
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

func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionregistrationv1beta1.AddToScheme(runtimeScheme)
	_ = v1.AddToScheme(runtimeScheme)
}
