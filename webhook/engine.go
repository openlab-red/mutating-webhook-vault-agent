package webhook

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"log"
	"os"
	"os/signal"
	"context"
	"time"
	"syscall"
	"github.com/spf13/viper"
)

func Start() {

	var engine = gin.Default()

	engine.GET("/health", Health)

	wk := WebHook{
		sidecarConfig: nil,
		server:        engine,
	}
	engine.GET("/mutate", wk.mutate)

	engine.RunTLS(":"+viper.GetString("port"), "/var/run/secrets/kubernetes.io/certs/tls.crt", "/var/run/secrets/kubernetes.io/certs/tls.key")

	shutdown(engine)
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
