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
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

func (wk *WebHook) mutate(context *gin.Context) {

	var admissionResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}

	if err := context.ShouldBindJSON(&admissionResponse); err == nil {
		log.Debugln(admissionResponse)
		admissionResponse = wk.admit(ar)
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

	log.WithFields(logrus.Fields{
		"Kind":           req.Kind,
		"Namespace":      req.Namespace,
		"Name":           req.Name,
		"UID":            req.UID,
		"PatchOperation": req.Operation,
		"UserInfo":       req.UserInfo,
	}).Infoln("AdmissionReview for")

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

func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionregistrationv1beta1.AddToScheme(runtimeScheme)
	_ = v1.AddToScheme(runtimeScheme)
}
