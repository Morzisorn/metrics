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
	buffer *bytes.Buffer
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

		c.Next()

		if c.Writer.Status() != 200 {
			return
		}

		if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			if strings.Contains(c.GetHeader("Content-Type"), "application/json") || strings.Contains(c.GetHeader("Content-Type"), "text/html") {
				buf := new(bytes.Buffer)
				c.Writer = &gzipResponseWriter{ResponseWriter: c.Writer, Writer: gzip.NewWriter(buf), buffer: buf}
				c.Writer.(*gzipResponseWriter).Close()

				c.Writer.Header().Set("Content-Encoding", "gzip")
				c.Writer.Write(buf.Bytes())
			}
		}
	}
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.Writer.Write(b)
}

func (g *gzipResponseWriter) Close() {
	if gz, ok := g.Writer.(*gzip.Writer); ok {
		gz.Close()
	}
}
