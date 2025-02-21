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

		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		buf := new(bytes.Buffer)
		gz := gzip.NewWriter(buf)
		defer gz.Close()

		gzw := &gzipResponseWriter{ResponseWriter: c.Writer, buffer: buf, Writer: gz}
		c.Writer = gzw
		defer gzw.Close()

		c.Next()

		if c.Writer.Status() != http.StatusOK {
			return
		}
		contentType := c.GetHeader("Content-Type")
		if !strings.Contains(contentType, "application/json") && !strings.Contains(contentType, "text/html") {
			c.Writer = gzw.ResponseWriter
			c.Writer.Write(buf.Bytes())
			return
		}

		c.Writer.Header().Set("Content-Encoding", "gzip")
		c.Writer.Write(buf.Bytes())
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
