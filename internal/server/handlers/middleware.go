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
	writer io.Writer
	buffer *bytes.Buffer
	status int
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

		gzw := &gzipResponseWriter{
			ResponseWriter: c.Writer,
			buffer:         buf,
			writer:         gz,
			status:         0,
		}

		c.Writer = gzw
		c.Next()

		/*
			if gzw.buffer.Len() == 0 {
				c.Writer.WriteHeader(gzw.status)
				gzw.Close()
				return
			}
		*/

		gzw.Close()

		contentType := c.Writer.Header().Get("Content-Type")

		if gzw.status < 200 || gzw.status >= 300 {
			if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html") {
				c.Writer.Header().Set("Content-Encoding", "gzip")
			}
			c.Writer = gzw.ResponseWriter
			c.Writer.WriteHeader(gzw.status)
			_, err := c.Writer.Write(buf.Bytes())
			if err != nil {
				logger.Log.Error("Error writing response", zap.Error(err))
			}
			return
		}

		if !(strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html")) {
			c.Writer = gzw.ResponseWriter
			c.Writer.WriteHeader(gzw.status)
			_, err := c.Writer.Write(buf.Bytes())
			if err != nil {
				logger.Log.Error("Error writing response", zap.Error(err))
			}
			return
		}

		c.Writer.Header().Set("Content-Encoding", "gzip")
		c.Writer.WriteHeader(gzw.status)

		gzw.Close()
		c.Writer = gzw.ResponseWriter
		_, err := c.Writer.Write(buf.Bytes())
		if err != nil {
			logger.Log.Error("Error writing response", zap.Error(err))
		}
	}
}
func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	if g.status == 0 {
		g.status = http.StatusOK
	}
	if gz, ok := g.writer.(*gzip.Writer); ok {
		return gz.Write(b)
	}
	return g.writer.Write(b)
}

func (g *gzipResponseWriter) Close() {
	if gz, ok := g.writer.(*gzip.Writer); ok {
		gz.Close()
	}
}

func (g *gzipResponseWriter) WriteHeader(code int) {
	g.status = code                    // Сохраняем код ответа
	g.ResponseWriter.WriteHeader(code) // Передаём оригинальному ResponseWriter
}
