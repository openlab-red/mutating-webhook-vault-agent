package webhook

import (
	"net/http"
	"github.com/spf13/viper"
)

func Handle() {

	wk := WebHook{
		sidecarConfig: nil,
	}

	http.HandleFunc("/mutate", wk.serve)

	server := &http.Server{
		Addr: ":" + viper.GetString("port"),
	}
	server.ListenAndServeTLS("/var/run/secrets/kubernetes.io/certs/tls.crt", "/var/run/secrets/kubernetes.io/certs/tls.key")

}
