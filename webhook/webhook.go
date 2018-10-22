package webhook

import (
	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/api/admission/v1beta1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/kubernetes/pkg/apis/core/v1"
	"net/http"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

type WebHook struct {
	sidecarConfig *Config
	server        *gin.Engine
}

type Config struct {
	Containers []corev1.Container `yaml:"containers"`
	Volumes    []corev1.Volume    `yaml:"volumes"`
}

func (wk *WebHook) mutate(context *gin.Context) {

	log.Println(context.Request.Method)
	var admissionResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}

	if err := context.BindJSON(&ar); err == nil {
		admissionResponse = wk.handle(&ar)
		admissionReview := v1beta1.AdmissionReview{}
		if admissionResponse != nil {
			admissionReview.Response = admissionResponse
			if ar.Request != nil {
				admissionReview.Response.UID = ar.Request.UID
			}
		}
		context.JSON(http.StatusOK, admissionReview)
	} else {
		admissionResponse = &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
		context.AbortWithStatusJSON(http.StatusBadRequest, admissionResponse)
	}

}

func (wk *WebHook) handle(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	log.Println(req)
	return nil
}

func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionregistrationv1beta1.AddToScheme(runtimeScheme)
	_ = v1.AddToScheme(runtimeScheme)
}
