package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	agent "github.com/morzisorn/metrics/internal/agent/services"
	"github.com/morzisorn/metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"resty.dev/v3"
)

func TestHTTPClient_SendMetric_Success(t *testing.T) {
	// Создаем тестовую метрику
	testMetric := &agent.Metric{
		Metric: models.Metric{
			ID:    "test_metric",
			MType: "gauge",
			Value: getPointer(42.5),
		},
	}

	// Создаем mock HTTP сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем HTTP метод
		assert.Equal(t, "POST", r.Method)

		// Проверяем URL path
		assert.Equal(t, "/update/", r.URL.Path)

		// Проверяем Content-Type header
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Проверяем тело запроса
		var receivedMetric agent.Metric
		err := json.NewDecoder(r.Body).Decode(&receivedMetric)
		require.NoError(t, err)

		assert.Equal(t, testMetric.ID, receivedMetric.ID)
		assert.Equal(t, testMetric.MType, receivedMetric.MType)
		assert.Equal(t, *testMetric.Value, *receivedMetric.Value)

		// Возвращаем успешный ответ
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Создаем HTTPClient с тестовым сервером
	serverHost := server.URL[7:] // "127.0.0.1:60254"
	client := &HTTPClient{
		BaseURL: serverHost,
		Client:  resty.New(),
	}

	// Выполняем тест
	err := client.SendMetric(testMetric)

	// Проверяем результат
	assert.NoError(t, err)
}

func TestHTTPClient_MetricSenderJob_Success(t *testing.T) {
	// Счетчик для отслеживания количества полученных метрик
	receivedMetrics := make([]agent.Metric, 0)
	var mutex sync.Mutex

	// Создаем mock HTTP сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем HTTP метод и path
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/update/", r.URL.Path)

		// Читаем и парсим метрику из запроса
		var receivedMetric agent.Metric
		err := json.NewDecoder(r.Body).Decode(&receivedMetric)
		assert.NoError(t, err)

		// Сохраняем полученную метрику (thread-safe)
		mutex.Lock()
		receivedMetrics = append(receivedMetrics, receivedMetric)
		mutex.Unlock()

		// Возвращаем успешный ответ
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Создаем HTTPClient с тестовым сервером
	serverHost := strings.TrimPrefix(server.URL, "http://")
	client := &HTTPClient{
		BaseURL: serverHost,
		Client:  resty.New(),
	}

	// Подготавливаем тестовые метрики
	testMetrics := []agent.Metric{
		{
			Metric: models.Metric{
				ID:    "cpu_usage",
				MType: "gauge",
				Value: getPointer(75.5),
			},
		},
		{
			Metric: models.Metric{
				ID:    "memory_usage",
				MType: "gauge",
				Value: getPointer(60.2),
			},
		},
		{
			Metric: models.Metric{
				ID:    "request_count",
				MType: "counter",
				Delta: getPointer(int64(42)),
			},
		},
	}

	// Создаем канал и отправляем в него метрики
	chIn := make(chan agent.Metric, len(testMetrics))
	for _, metric := range testMetrics {
		chIn <- metric
	}
	close(chIn) // Закрываем канал, чтобы горутина завершилась

	// Создаем WaitGroup для ожидания завершения горутины
	var wg sync.WaitGroup
	wg.Add(1)

	// Запускаем metricSenderJob в горутине
	go metricSenderJob(client, chIn, &wg)

	// Ждем завершения горутины
	wg.Wait()

	// Проверяем результаты
	mutex.Lock()
	assert.Len(t, receivedMetrics, len(testMetrics), "Should receive all sent metrics")

	// Проверяем, что все метрики были получены корректно
	for i, expectedMetric := range testMetrics {
		found := false
		for _, receivedMetric := range receivedMetrics {
			if receivedMetric.ID == expectedMetric.Metric.ID {
				assert.Equal(t, expectedMetric.Metric.MType, receivedMetric.MType)
				if expectedMetric.Metric.Value != nil {
					assert.Equal(t, *expectedMetric.Metric.Value, *receivedMetric.Value)
				}
				if expectedMetric.Metric.Delta != nil {
					assert.Equal(t, *expectedMetric.Metric.Delta, *receivedMetric.Delta)
				}
				found = true
				break
			}
		}
		assert.True(t, found, "Metric %s should be found in received metrics", expectedMetric.Metric.ID)
		_ = i // использование переменной для избежания warning
	}
	mutex.Unlock()
}

func TestHTTPClient_SendMetricsBatch_Success(t *testing.T) {
	// Переменная для хранения полученного батча метрик
	var receivedBatch []models.Metric

	// Создаем mock HTTP сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем HTTP метод и path
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/updates/", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Читаем и парсим батч метрик из запроса
		err := json.NewDecoder(r.Body).Decode(&receivedBatch)
		assert.NoError(t, err)

		// Возвращаем успешный ответ
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Создаем HTTPClient с тестовым сервером
	serverHost := strings.TrimPrefix(server.URL, "http://")
	client := &HTTPClient{
		BaseURL: serverHost,
		Client:  resty.New(),
	}

	// Подготавливаем тестовые метрики
	testMetrics := map[string]agent.Metric{
		"cpu_usage": {
			Metric: models.Metric{
				ID:    "cpu_usage",
				MType: "gauge",
				Value: getPointer(75.5),
			},
		},
		"memory_usage": {
			Metric: models.Metric{
				ID:    "memory_usage",
				MType: "gauge",
				Value: getPointer(60.2),
			},
		},
		"request_count": {
			Metric: models.Metric{
				ID:    "request_count",
				MType: "counter",
				Delta: getPointer(int64(42)),
			},
		},
	}

	// Создаем реальный объект agent.Metrics
	metricsObj := &agent.Metrics{
		Metrics: testMetrics,
		Mu:      sync.RWMutex{},
	}

	// Выполняем тест
	err := client.SendMetricsBatch(metricsObj)

	// Проверяем результат
	require.NoError(t, err)

	// Проверяем, что получили правильное количество метрик в батче
	assert.Len(t, receivedBatch, len(testMetrics), "Should receive all metrics in batch")

	// Проверяем, что все метрики были получены корректно
	for _, expectedMetric := range testMetrics {
		found := false
		for _, receivedMetric := range receivedBatch {
			if receivedMetric.ID == expectedMetric.Metric.ID {
				assert.Equal(t, expectedMetric.Metric.MType, receivedMetric.MType)
				if expectedMetric.Metric.Value != nil {
					assert.Equal(t, *expectedMetric.Metric.Value, *receivedMetric.Value)
				}
				if expectedMetric.Metric.Delta != nil {
					assert.Equal(t, *expectedMetric.Metric.Delta, *receivedMetric.Delta)
				}
				found = true
				break
			}
		}
		assert.True(t, found, "Metric %s should be found in received batch", expectedMetric.Metric.ID)
	}

	// Проверяем, что ResetCounter был вызван (косвенно через проверку состояния)
	// В реальном тесте можно было бы проверить, что счетчики сброшены
}

// Вспомогательная функция для создания указателя
func getPointer[T any](v T) *T {
	return &v
}
