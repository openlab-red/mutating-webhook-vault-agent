package webhook

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"os/signal"
	"context"
	"time"
	"syscall"
	"github.com/spf13/viper"
)


func Start() {
	var engine = gin.New()

	InitLogrus(engine)

	engine.GET("/health", Health)

	webhook(engine)

	engine.RunTLS(":"+viper.GetString("port"), "/var/run/secrets/kubernetes.io/certs/tls.crt", "/var/run/secrets/kubernetes.io/certs/tls.key")
	shutdown(engine)
}

func webhook(engine *gin.Engine) {
	sidecarConfig, err := LoadConfig("/var/run/secrets/kubernetes.io/config/sidecarconfig.yaml")
	if err != nil {
		log.Errorln(err)
	}
	wk := WebHook{
		sidecarConfig: sidecarConfig,
	}
	engine.POST("/mutate", wk.mutate)
}

func shutdown(engine *gin.Engine) {
	srv := &http.Server{
		Addr:    viper.GetString("port"),
		Handler: engine,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
}
