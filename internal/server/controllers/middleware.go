package controllers

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

	"compress/gzip"

	"github.com/gin-gonic/gin"
	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/server/logger"
	"go.uber.org/zap"
)

type gzipResponseWriter struct {
	gin.ResponseWriter
	writer io.Writer
	buffer *bytes.Buffer
	status int
}

type responseWriter struct {
	gin.ResponseWriter
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
	g.status = code
	g.ResponseWriter.WriteHeader(code)
}

func SignMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.GetService("server").Config.Key == "" {
			return
		}

		hashReq := c.GetHeader("HashSHA256")
		if hashReq == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "must have HashSHA256 header"})
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		fmt.Println("body :", string(body))
		if err != nil {
			logger.Log.Error("Error reading body", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "error reading body"})
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		hashServer := getHash(body)
		decHashReq, err := hex.DecodeString(hashReq)

		if string(decHashReq) != string(hashServer[:]) {
			logger.Log.Error("Incorrect sign hash", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "incorrect sign hash"})
			return
		}

		rw := &responseWriter{
			ResponseWriter: c.Writer,
			buffer:         bytes.NewBuffer(nil),
			status:         0,
		}

		c.Writer = rw

		c.Next()

		if config.GetService("server").Config.Key == "" {
			return
		}

		body = []byte(rw.buffer.String())

		hash := getHash(body)
		hashHex := hex.EncodeToString(hash[:])

		c.Writer.Header().Set("HashSHA256", hashHex)
		c.Writer.WriteHeader(rw.status)

		c.Writer = rw.ResponseWriter

		_, err = c.Writer.Write(rw.buffer.Bytes())
		if err != nil {
			logger.Log.Error("Error writing response", zap.Error(err))
		}
		return

	}
}

func getHash(body []byte) [32]byte {
	service := config.GetService("server")
	str := append(body, []byte(service.Config.Key)...)
	return sha256.Sum256(str)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	return rw.buffer.Write(b)
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
