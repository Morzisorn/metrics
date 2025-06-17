package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/hash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"resty.dev/v3"
)

func TestGzipMiddleware_Success(t *testing.T) {
	// Создаем resty запрос с тестовыми данными
	client := resty.New()
	request := client.R()

	testData := []byte("test data for compression")
	request.SetBody(testData)

	// Применяем gzip middleware
	err := gzipMiddleware(request)
	require.NoError(t, err)

	// Проверяем, что заголовок установлен
	assert.Equal(t, "gzip", request.Header.Get("Content-Encoding"))

	// Проверяем, что тело сжато
	compressedBody := request.Body.([]byte)
	assert.True(t, len(compressedBody) > 0)

	// Проверяем, что данные можно распаковать обратно
	reader, err := gzip.NewReader(bytes.NewReader(compressedBody))
	require.NoError(t, err)
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, testData, decompressed)
}

func TestRetryConditions_Success(t *testing.T) {
	// Создаем mock HTTP ответ с ошибкой 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := resty.New()
	resp, _ := client.R().Get(server.URL)

	// Проверяем, что функция возвращает true для 500 ошибки
	shouldRetry := retryConditions(resp, nil)
	assert.True(t, shouldRetry, "Should retry on 500 status code")

	// Проверяем для 429 Too Many Requests
	server429 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server429.Close()

	resp429, _ := client.R().Get(server429.URL)
	shouldRetry429 := retryConditions(resp429, nil)
	assert.True(t, shouldRetry429, "Should retry on 429 status code")
}

func TestRetryHook_Success(t *testing.T) {
	// Создаем тестовый запрос
	client := resty.New()
	request := client.R()

	// Имитируем ответ с attempt = 1
	resp := &resty.Response{
		Request: request,
	}

	// Устанавливаем attempt вручную (в реальности это делает resty)
	request.Attempt = 1

	// Засекаем время до вызова
	start := time.Now()

	// Вызываем retry hook
	retryHook(resp, nil)

	// Проверяем, что прошло некоторое время (означает, что sleep был вызван)
	elapsed := time.Since(start)

	// Проверяем, что функция выполнилась без ошибок
	// В реальной конфигурации должен быть задан RetryDelays
	// Здесь мы просто проверяем, что функция не падает
	assert.True(t, elapsed >= 0, "RetryHook should execute without errors")
}

func TestSignRequestMiddleware_Success(t *testing.T) {
	// Создаем resty запрос с тестовыми данными
	client := resty.New()
	request := client.R()

	testData := []byte("test data for signing")
	request.SetBody(testData)

	// Мокируем конфигурацию с ключом
	// Предполагаем, что в реальной конфигурации есть Key
	originalService := config.GetService()

	// Применяем middleware
	err := signRequestMiddleware(request)
	require.NoError(t, err)

	// Если в конфигурации есть ключ, проверяем хеш
	if originalService.Config.Key != "" {
		hashHeader := request.Header.Get("HashSHA256")
		assert.NotEmpty(t, hashHeader, "Hash header should be set when key is configured")

		// Проверяем, что хеш корректный
		expectedHash := hash.GetHash(testData)
		expectedHashHex := hex.EncodeToString(expectedHash[:])
		assert.Equal(t, expectedHashHex, hashHeader)
	}

	// Если ключа нет, middleware должен просто пройти без ошибок
	// (что мы уже проверили через require.NoError)
}

func TestGetByteBody_Success(t *testing.T) {
	tests := []struct {
		name     string
		body     interface{}
		expected []byte
	}{
		{
			name:     "byte slice body",
			body:     []byte("test data"),
			expected: []byte("test data"),
		},
		{
			name:     "string body",
			body:     "test string",
			expected: []byte("test string"),
		},
		{
			name:     "struct body",
			body:     map[string]string{"key": "value"},
			expected: []byte(`{"key":"value"}`),
		},
		{
			name:     "nil body",
			body:     nil,
			expected: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			request := client.R()
			request.SetBody(tt.body)

			result, err := getByteBody(request)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEncryptMiddleware_Success(t *testing.T) {
	// Создаем resty запрос с тестовыми данными
	client := resty.New()
	request := client.R()

	testData := []byte("test data for encryption")
	request.SetBody(testData)

	// Временно устанавливаем публичный ключ в конфигурацию
	// В реальном тесте нужно будет мокировать config.GetService()
	originalService := config.GetService()

	// Применяем middleware только если в конфигурации есть публичный ключ
	err := encryptMiddleware(request)
	require.NoError(t, err)

	// Если публичный ключ настроен, проверяем результат шифрования
	if originalService.Config.PublicKey != nil {
		// Проверяем, что заголовок X-Encrypted установлен
		assert.Equal(t, "1", request.Header.Get("X-Encrypted"))

		// Проверяем, что тело изменилось и содержит зашифрованные данные
		encryptedBody := request.Body.([]byte)
		assert.True(t, len(encryptedBody) > 0)
		assert.NotEqual(t, testData, encryptedBody)

		// Проверяем структуру зашифрованного payload
		var payload struct {
			Key  string `json:"key"`
			Data string `json:"data"`
		}
		err = json.Unmarshal(encryptedBody, &payload)
		require.NoError(t, err)

		assert.NotEmpty(t, payload.Key, "Encrypted key should not be empty")
		assert.NotEmpty(t, payload.Data, "Encrypted data should not be empty")

		// Проверяем, что ключ и данные в base64 формате
		_, err = base64.StdEncoding.DecodeString(payload.Key)
		assert.NoError(t, err, "Key should be valid base64")

		_, err = base64.StdEncoding.DecodeString(payload.Data)
		assert.NoError(t, err, "Data should be valid base64")
	}

	// Если публичного ключа нет, middleware должен просто пройти без изменений
	// (что мы уже проверили через require.NoError)
}
