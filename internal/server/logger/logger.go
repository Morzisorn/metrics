// Package logger i used to log info
package logger

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Log *zap.Logger = zap.NewNop()

// Init creates new logger
// Should be used once when app starts
func Init() error {
	cfg := zap.NewProductionConfig()

	cfg.Level.SetLevel(zap.InfoLevel)

	cfg.EncoderConfig.StacktraceKey = ""

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl
	return nil
}

// LoggerMiddleware logs URI and method of request
// and more field of response depends on status code
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		Log.Info("Request started",
			zap.String("URI", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
		)

		c.Next()

		duration := time.Since(start)

		if c.Writer.Status() == 200 {
			Log.Info("Request completed",
				zap.String("duration", duration.String()),
				zap.String("status", strconv.Itoa(c.Writer.Status())),
				zap.String("size", strconv.Itoa(c.Writer.Size())),
			)
		} else {
			Log.Error("Request failed",
				zap.String("URI", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("status", strconv.Itoa(c.Writer.Status())),
				zap.String("size", strconv.Itoa(c.Writer.Size())),
				zap.String("body", c.Errors.String()),
				zap.String("duration", duration.String()),
			)
		}
	}
}
