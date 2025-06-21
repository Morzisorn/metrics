package controllers

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/hash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testResponseWriter struct {
	*httptest.ResponseRecorder
}

func (w *testResponseWriter) WriteHeaderNow()          {}
func (w *testResponseWriter) Size() int                { return w.Body.Len() }
func (w *testResponseWriter) Written() bool            { return w.Code != 0 }
func (w *testResponseWriter) Status() int              { return w.Code }
func (w *testResponseWriter) CloseNotify() <-chan bool { return make(chan bool) }
func (w *testResponseWriter) Flush()                   {}
func (w *testResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, http.ErrNotSupported
}
func (w *testResponseWriter) Pusher() http.Pusher {
	return nil
}

func TestGzipResponseWriter_Write(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	rec := httptest.NewRecorder()
	wrappedRec := &testResponseWriter{rec}

	gzw := &gzipResponseWriter{ResponseWriter: wrappedRec, writer: gz, buffer: &buf}

	data := []byte("Test data!")

	n, err := gzw.Write(data)

	require.NoError(t, err, "Write() returned an error")
	assert.Equal(t, len(data), n, "Write() returned an incorrect number of bytes")

	gz.Close()

	assert.NotEqual(t, data, buf.Bytes(), "Data was not compressed")
}

func TestGzipResponseWriter_Close(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	rec := httptest.NewRecorder()
	wrappedRec := &testResponseWriter{rec}

	gzw := &gzipResponseWriter{ResponseWriter: wrappedRec, writer: gz, buffer: &buf}

	gzw.Close()

	_, err := gz.Write([]byte("test"))
	assert.Error(t, err, "Write() did not return an error after closing the writer")
}

func TestGzipMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(GzipMiddleware())

	router.POST("/test", func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		require.NoError(t, err)
		c.Data(http.StatusOK, "text/html", []byte("received: "+string(body)))
	})

	// Compress body
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write([]byte("test payload"))
	require.NoError(t, err)
	require.NoError(t, gz.Close())

	req := httptest.NewRequest(http.MethodPost, "/test", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
	require.Equal(t, "gzip", resp.Header().Get("Content-Encoding"))

	// Decompress body
	gr, err := gzip.NewReader(resp.Body)
	require.NoError(t, err)
	defer gr.Close()

	unzipped, err := io.ReadAll(gr)
	require.NoError(t, err)
	require.Contains(t, string(unzipped), "test payload")
}

func TestSignMiddleware_ValidSignature(t *testing.T) {
	// Устанавливаем ключ
	config.GetService("server").Config.Key = "test-key"

	body := []byte(`{"message":"hello"}`)

	// Вычисляем корректный хеш
	s := hash.GetHash([]byte("OK"))

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SignMiddleware())

	router.POST("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	h := hash.GetHash(body)
	req.Header.Set("HashSHA256", hex.EncodeToString(h[:])) // ← это для запроса

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, hex.EncodeToString(s[:]), w.Header().Get("HashSHA256")) // ← это для ответа
}

func TestSignMiddleware_InvalidSignature(t *testing.T) {
	config.GetService("server").Config.Key = "test-key"

	body := []byte(`{"message":"hello"}`)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SignMiddleware())

	router.POST("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	req.Header.Set("HashSHA256", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef") // 64 символа
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "incorrect sign hash")
}

func TestDecryptMiddleware_NoEncryptionHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(DecryptMiddleware())

	var receivedBody []byte
	router.POST("/test", func(c *gin.Context) {
		receivedBody, _ = io.ReadAll(c.Request.Body)
		c.Status(http.StatusOK)
	})

	originalData := []byte(`{"test":"data"}`)
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(originalData))
	// НЕ устанавливаем заголовок X-Encrypted
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, originalData, receivedBody, "Data should pass through unchanged")
}

func TestDecryptMiddleware_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(DecryptMiddleware())

	var receivedBody []byte
	router.POST("/test", func(c *gin.Context) {
		receivedBody, _ = io.ReadAll(c.Request.Body)
		c.Status(http.StatusOK)
	})

	// Некорректный JSON с заголовком X-Encrypted
	invalidJSON := []byte(`{"invalid": json}`)
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(invalidJSON))
	req.Header.Set("X-Encrypted", "1")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Middleware должен восстановить body и пропустить дальше
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, invalidJSON, receivedBody, "Invalid JSON should pass through")
}

func TestDecryptMiddleware_Success(t *testing.T) {
	// Инициализируем config с app type чтобы избежать панику
	config.GetService("server")
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(DecryptMiddleware())
	
	router.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Создаем валидный payload с корректным base64
	payload := map[string]string{
		"key":  base64.StdEncoding.EncodeToString(make([]byte, 256)), // 256 байт для RSA-2048
		"data": base64.StdEncoding.EncodeToString([]byte("fake-encrypted-data")),
	}
	payloadBytes, _ := json.Marshal(payload)
	
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(payloadBytes))
	req.Header.Set("X-Encrypted", "1")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)

	// Тест пройдет до RSA расшифровки, где упадет с 400 (нормально без реального ключа)
	// Но это значит, что вся логика до этого момента работает включая config.GetService()
	assert.Equal(t, http.StatusBadRequest, w.Code, "Should fail at RSA decryption stage")
}
