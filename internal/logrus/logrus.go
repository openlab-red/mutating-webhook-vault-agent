package logrus

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var log = logrus.New()

// GIN placeholder for GIN logs
const GIN = "[GIN]"

// InitLogrus initialise the log from LogRus
func InitLogrus(engine *gin.Engine) {
	level, err := logrus.ParseLevel(viper.GetString("log-level"))
	if err != nil {
		log.Fatalln(err)
	} else {
		log.Level = level
	}
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	engine.Use(LoggerWithLogrus(log))

}

// Log returns log instance
func Log() *logrus.Logger {
	return log
}

// LoggerWithLogrus configure GIN to work with logrus
func LoggerWithLogrus(log *logrus.Logger) gin.HandlerFunc {

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)

		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		comment := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if raw != "" {
			path = path + "?" + raw
		}

		requestLogger := log.WithFields(logrus.Fields{
			"statusCode": statusCode,
			"latency":    latency,
			"clientIP":   clientIP,
			"method":     method,
			"path":       path,
			"user-agent": c.Request.UserAgent(),
			"comment":    comment,
		})
		if statusCode > 400 || len(c.Errors) > 0 {
			requestLogger.Error(GIN, path)
		} else {
			requestLogger.Debug(GIN, path)
		}
	}
}
