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
	"github.com/sirupsen/logrus"
	"github.com/gin-gonic/contrib/ginrus"
)

var log = logrus.New()

func Start() {
	var engine = gin.Default()

	initLog(engine)

	engine.GET("/health", Health)

	wk := WebHook{
		sidecarConfig: nil,
		server:        engine,
	}
	engine.GET("/mutate", gin.WrapF(wk.handler))

	engine.RunTLS(":"+viper.GetString("port"), "/var/run/secrets/kubernetes.io/certs/tls.crt", "/var/run/secrets/kubernetes.io/certs/tls.key")

	shutdown(engine)
}

func initLog(engine *gin.Engine) {
	level, err := logrus.ParseLevel(viper.GetString("log-level"))
	if err != nil {
		log.Fatalln(err)
	} else {
		log.Level = level
	}

	engine.Use(ginrus.Ginrus(log, time.RFC3339, true))

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
