package rest

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"

	"compress/gzip"

	"github.com/gin-gonic/gin"
	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/hash"
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

var gzipWriterPool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(io.Discard)
	},
}

// GzipMiddleware decompresses requests and compresses responses bases on
// request headers.
func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
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
		}

		if !strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		buf := new(bytes.Buffer)
		gz := gzipWriterPool.Get().(*gzip.Writer)
		gz.Reset(buf)

		gzw := &gzipResponseWriter{
			ResponseWriter: c.Writer,
			buffer:         buf,
			writer:         gz,
			status:         0,
		}

		c.Writer = gzw
		c.Next()

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
		gzipWriterPool.Put(gz)
	}
}

func (g *gzipResponseWriter) WriteHeader(code int) {
	g.status = code
	g.ResponseWriter.WriteHeader(code)
}

// SignMiddleware checks if key is set up.
// If yes, it checks if HashSHA256 header is correct.
func SignMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.GetService().Config.Key == "" {
			logger.Log.Info("Skipping middleware: no key configured")
			c.Next()
			return
		}

		hashReq := c.Request.Header.Get("HashSHA256")

		if hashReq == "" {
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
		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		hashServer := hash.GetHash(body)
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

		hash := hash.GetHash(rw.buffer.Bytes())
		c.Writer.Header().Set("HashSHA256", hex.EncodeToString(hash[:]))
		c.Writer.WriteHeader(rw.status)

		c.Writer = rw.ResponseWriter
		if _, err := c.Writer.Write(rw.buffer.Bytes()); err != nil {
			logger.Log.Error("Error writing response", zap.Error(err))
		}
	}
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

func DecryptMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		service := config.GetService()
		if service.Config.PrivateKey == nil {
			c.Next()
			return
		}

		if c.GetHeader("X-Encrypted") != "1" {
			c.Next()
			return
		}

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		var w struct {
			Key  string `json:"key"`
			Data string `json:"data"`
		}
		if err := json.Unmarshal(bodyBytes, &w); err != nil {
			c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			c.Next()
			return
		}

		encKey, err := base64.StdEncoding.DecodeString(w.Key)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		aesKey, err := rsa.DecryptOAEP(
			sha256.New(),
			rand.Reader,
			service.Config.PrivateKey,
			encKey,
			nil,
		)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		encData, err := base64.StdEncoding.DecodeString(w.Data)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		block, err := aes.NewCipher(aesKey)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		gcm, err := cipher.NewGCM(block)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		nonceSize := gcm.NonceSize()
		if len(encData) < nonceSize {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		nonce, ciphertext := encData[:nonceSize], encData[nonceSize:]

		plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewReader(plaintext))
		c.Request.Header.Del("X-Encrypted")
		c.Next()
	}
}

func ValidateIPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cnfg := config.GetService()
		if cnfg.Config.TrustedSubnet == "" {
			c.Next()
			return
		}

		realIP := c.Request.Header.Get("X-Real-IP")
		if realIP == "" {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		trusted, err := isIPInSubnet(realIP, cnfg.Config.TrustedSubnet)
		if err != nil {
			logger.Log.Error("Check is IP trusted error", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		
		if !trusted {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}

func isIPInSubnet(ipStr, cidr string) (bool, error) {
	// Parse CIDR
	_, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, fmt.Errorf("invalid CIDR: %v", err)
	}

	// Parse IP
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	return subnet.Contains(ip), nil
}
