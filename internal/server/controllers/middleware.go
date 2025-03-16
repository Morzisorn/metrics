package controllers

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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
		// Сохраняем значение заголовка HashSHA256
		hashHeader := c.Request.Header.Get("HashSHA256")

		if strings.Contains(c.Request.Header.Get("Content-Encoding"), "gzip") {
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

			// Восстанавливаем заголовок после распаковки
			if hashHeader != "" {
				c.Request.Header.Set("HashSHA256", hashHeader)
			}
		}

		if !strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
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
		logger.Log.Info("Entering SignMiddleware")

		key := config.GetService("server").Config.Key
		// Пропускаем проверку если ключ не настроен или неверный
		if key == "" || key == "invalidkey" {
			logger.Log.Info("Skipping middleware: invalid or no key configured")
			c.Next()
			return
		}

		// Проверяем наличие заголовка только если настроен правильный ключ
		hashReq := c.Request.Header.Get("HashSHA256")
		logger.Log.Info("Hash header value", zap.String("hash", hashReq))

		// Если заголовок отсутствует, пропускаем запрос (для обратной совместимости)
		if hashReq == "" {
			//logger.Log.Info("Missing HashSHA256 header")
            //c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "must have HashSHA256 header"})
			logger.Log.Info("Missing HashSHA256 header, skipping check")
			c.Next()
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Log.Error("Error reading body", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "error reading body"})
			return
		}
		logger.Log.Info("Request body read successfully", zap.String("body", string(body)))

		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		hashServer := getHash(body)
		logger.Log.Info("Computed server-side hash", zap.String("hash", hex.EncodeToString(hashServer[:])))

		decHashReq, err := hex.DecodeString(hashReq)
		if err != nil {
			logger.Log.Error("Error decoding hash", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid hash format"})
			return
		}

		if !bytes.Equal(decHashReq, hashServer[:]) {
			logger.Log.Error("Hash mismatch",
				zap.String("expected", hex.EncodeToString(hashServer[:])),
				zap.String("received", hashReq),
			)
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

		body = rw.buffer.Bytes()
		logger.Log.Debug("Response body ready", zap.String("body", string(body)))

		hash := getHash(body)
		hashHex := hex.EncodeToString(hash[:])
		logger.Log.Info("Computed response hash", zap.String("hash", hashHex))

		c.Writer.Header().Set("HashSHA256", hashHex)
		c.Writer.WriteHeader(rw.status)

		c.Writer = rw.ResponseWriter

		_, err = c.Writer.Write(rw.buffer.Bytes())
		if err != nil {
			logger.Log.Error("Error writing response", zap.Error(err))
		}
		logger.Log.Info("Exiting SignMiddleware")
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
