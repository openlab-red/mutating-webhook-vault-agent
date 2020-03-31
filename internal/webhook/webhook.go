package webhook

import (
	"net/http"

	"github.com/gin-gonic/gin"
	logger "github.com/openlab-red/mutating-webhook-vault-agent/internal/logrus"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
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
		{"sidecar.agent.vaultproject.io/secret", alwaysValidFunc},
		{"sidecar.agent.vaultproject.io/filename", alwaysValidFunc},
		{"sidecar.agent.vaultproject.io/role", alwaysValidFunc},
	}

	annotationPolicy        = annotationRegistry[0]
	annotationStatus        = annotationRegistry[1]
	annotationSecret        = annotationRegistry[2]
	annotationVaultFileName = annotationRegistry[3]
	annotationVaultRole     = annotationRegistry[4]

	ignoredNamespaces = []string{
		metav1.NamespaceSystem,
		metav1.NamespacePublic,
	}

	log = logger.Log()
)

// Mutate AdmissionReview Request
func (wk *WebHook) Mutate(context *gin.Context) {

	var admissionReview v1.AdmissionReview

	if err := context.ShouldBindJSON(&admissionReview); err == nil {
		log.WithFields(logrus.Fields{
			"AdmissionReview": admissionReview,
		}).Debugln("AdmissionReview: ")
		admissionReview.Response = wk.admit(admissionReview)
		log.WithFields(logrus.Fields{
			"AdmissionResponse": admissionReview.Response,
		}).Debugln("AdmissionReview: ")
		context.JSON(http.StatusOK, &admissionReview)
	} else {
		log.WithFields(logrus.Fields{
			"Context": context,
			"Error":   err,
		}).Errorln("Mutate Request: ")
		context.AbortWithStatusJSON(http.StatusBadRequest, ToAdmissionResponseError(err))
	}

}

func (wk *WebHook) admit(ar v1.AdmissionReview) *v1.AdmissionResponse {
	req := ar.Request
	pod := corev1.Pod{}
	var err error
	var name string

	if err = Pod(req.Object.Raw, &pod); err != nil {
		return ToAdmissionResponseError(err)
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

	if !isRequired(ignoredNamespaces, &pod) {
		log.WithFields(logrus.Fields{
			"Kind":           req.Kind,
			"Namespace":      req.Namespace,
			"Name":           pod.Name,
			"UID":            req.UID,
			"PatchOperation": req.Operation,
			"UserInfo":       req.UserInfo,
		}).Infoln("Admission Not Required")
		return &v1.AdmissionResponse{
			Allowed: true,
			UID:     req.UID,
		}
	}

	//sidecar data
	name, err = GetDeploymentName(pod.OwnerReferences[0].Name)
	if err != nil {
		return ToAdmissionResponseError(err)
	}

	data := SidecarData{
		Name:          name,
		Container:     pod.Spec.Containers[0],
		TokenVolume:   FindTokenVolumeName(pod.Spec.Volumes),
		VaultSecret:   GetAnnotationValue(pod, annotationSecret, ""),
		VaultFileName: GetAnnotationValue(pod, annotationVaultFileName, "application.yaml"),
		VaultRole:     GetAnnotationValue(pod, annotationVaultRole, "example"),
	}

	// agent configMap
	_, err = agentConfigMap(VaultAgentConfigPrefix, pod, wk, &data, false)
	if err != nil {
		return ToAdmissionResponseError(err)
	}

	// ca-bundle
	_, err = caBundleConfigMap(pod, wk, &data)
	if err != nil {
		return ToAdmissionResponseError(err)
	}

	wk.VaultConfig, err = inject(&data, wk.SidecarConfig)
	if err != nil {
		return ToAdmissionResponseError(err)
	}
	annotations := map[string]string{annotationStatus.name: "injected"}

	//patch
	patches, err := CreatePatch(&pod, wk.VaultConfig, annotations)
	if err != nil {
		return ToAdmissionResponseError(err)
	}

	log.Debugf("AdmissionResponse: patch=%v\n", string(patches))

	log.WithFields(logrus.Fields{
		"Kind":           req.Kind,
		"Namespace":      req.Namespace,
		"Name":           pod.Name,
		"UID":            req.UID,
		"PatchOperation": req.Operation,
		"UserInfo":       req.UserInfo,
	}).Infoln("AdmissionResponse Allowed for")

	return &v1.AdmissionResponse{
		Allowed: true,
		UID:     req.UID,
		Patch:   patches,
		PatchType: func() *v1.PatchType {
			pt := v1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}
