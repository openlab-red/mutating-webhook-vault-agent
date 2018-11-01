package engine

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"io/ioutil"
	"crypto/sha256"
	"github.com/ghodss/yaml"
	"github.com/openlab-red/mutating-webhook-vault-agent/pkg/kubernetes"
)

func Start() {
	var engine = gin.New()

	InitLogrus(engine)

	engine.GET("/health", health)

	hook(engine)

	engine.RunTLS(":"+viper.GetString("port"), "/var/run/secrets/kubernetes.io/certs/tls.crt", "/var/run/secrets/kubernetes.io/certs/tls.key")
	shutdown(engine)
}

func hook(engine *gin.Engine) {

	sideCarFile := kubernetes.Config{}
	load("/var/run/secrets/kubernetes.io/config/sidecarconfig.yaml", &sideCarFile)

	wk := kubernetes.WebHook{
		Config: &sideCarFile,
	}

	engine.POST("/mutate", wk.Mutate)

}

func load(file string, c interface{}) {

	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalln(err)
	}

	if err := yaml.Unmarshal(data, c); err != nil {
		log.Warnf("Failed to parse %s", string(data))
	}

	log.Debugf("New configuration: sha256sum %x", sha256.Sum256(data))
	log.Debugf("Config: %v", c)
}
