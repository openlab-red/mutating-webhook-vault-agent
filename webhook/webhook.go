package webhook

import (
	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/api/admission/v1beta1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/kubernetes/pkg/apis/core/v1"
	"net/http"
	"github.com/sirupsen/logrus"
	"strings"
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
		{"sidecar.agent..vaultproject.io/status", alwaysValidFunc},
	}

	annotationPolicy = annotationRegistry[0]
	annotationStatus = annotationRegistry[1]
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

	if err := Pod(req.Object.Raw, pod); err != nil {
		return ToAdmissionResponse(err)
	}

	PotentialPodAndNamespace(req, &pod)

	log.WithFields(logrus.Fields{
		"Kind":           req.Kind,
		"Namespace":      req.Namespace,
		"Name":           req.Name,
		"UID":            req.UID,
		"PatchOperation": req.Operation,
		"UserInfo":       req.UserInfo,
	}).Infoln("AdmissionReview for")

	if !injectionStatus(&pod) {
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	patches, err := CreatePatch(&pod, wk.sidecarConfig, nil)

	if err != nil {
		return ToAdmissionResponse(err)
	}

	return &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patches,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

func injectionStatus(pod *corev1.Pod) bool {

	required := false
	metadata := pod.ObjectMeta

	if metadata.Annotations != nil {
		status := metadata.Annotations[annotationStatus.name]

		if strings.ToLower(status) == "injected" {
			required = false
		} else {
			switch strings.ToLower(metadata.Annotations[annotationPolicy.name]) {
			default:
				required = false
			case "true":
				required = true
			}
		}

		log.WithFields(logrus.Fields{
			"name":      metadata.Name,
			"namespace": metadata.Namespace,
			"status":    status,
			"required":  required,
		}).Infoln("Mutation policy for")
	}

	return required
}

func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionregistrationv1beta1.AddToScheme(runtimeScheme)
	_ = v1.AddToScheme(runtimeScheme)
}
