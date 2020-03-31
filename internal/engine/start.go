package engine

import (
	"github.com/gin-gonic/gin"
	"github.com/openlab-red/mutating-webhook-vault-agent/internal/logrus"
	"github.com/openlab-red/mutating-webhook-vault-agent/internal/webhook"
	"github.com/spf13/viper"
)

// Start GIN Server Engine
func Start() {
	var engine = gin.New()

	logrus.InitLogrus(engine)

	engine.GET("/health", health)

	hook(engine)

	engine.RunTLS(":"+viper.GetString("port"), "/var/run/secrets/kubernetes.io/certs/tls.crt", "/var/run/secrets/kubernetes.io/certs/tls.key")

	shutdown(engine)
}

func hook(engine *gin.Engine) {

	sidecarConfig := webhook.SidecarConfig{}
	webhook.Load("/var/run/secrets/kubernetes.io/config/sidecarconfig.yaml", &sidecarConfig)

	wk := webhook.WebHook{
		SidecarConfig: &sidecarConfig,
	}

	engine.POST("/mutate", wk.Mutate)

}
