package server

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"compress/gzip"

	"github.com/gin-gonic/gin"
	"github.com/morzisorn/metrics/internal/server/logger"
	"go.uber.org/zap"
)

type gzipResponseWriter struct {
	gin.ResponseWriter
	Writer io.Writer
}

func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				logger.Log.Error("Error reading gzip body", zap.Error(err))
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "error reading gzip body"})
				return
			}
			defer gz.Close()

			body, err := io.ReadAll(gz)
			if err != nil {
				logger.Log.Error("Error reading gzip body", zap.Error(err))
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "error reading gzip body"})
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
		}

		buf := new(bytes.Buffer)
		gz := gzip.NewWriter(buf)
		defer gz.Close()

		gzw := &gzipResponseWriter{ResponseWriter: c.Writer, Writer: gz}
		c.Writer = gzw

		c.Next()

		if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			if strings.Contains(c.GetHeader("Accept-Content"), "application/json") || strings.Contains(c.GetHeader("Accept-Content"), "text/html") {
				c.Header("Content-Encoding", "gzip")
				gz.Close()
			}
		}
	}
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.Writer.Write(b)
}
