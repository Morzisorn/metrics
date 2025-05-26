package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitMetrics(t *testing.T) {
	baseValue := 1.5
	baseDelta := int64(10)
	j := 3

	metrics := initMetrics(baseValue, baseDelta, j)

	require.Len(t, metrics, 3)

	// Проверка RandomValue
	require.Equal(t, "RandomValue", metrics[0].ID)
	require.Equal(t, "gauge", metrics[0].Type)
	require.NotNil(t, metrics[0].Value)
	require.Equal(t, baseValue+float64(j), *metrics[0].Value)

	// Проверка PollCount
	require.Equal(t, "PollCount", metrics[1].ID)
	require.Equal(t, "counter", metrics[1].Type)
	require.NotNil(t, metrics[1].Delta)
	require.Equal(t, baseDelta+int64(j), *metrics[1].Delta)

	// Проверка Alloc
	require.Equal(t, "Alloc", metrics[2].ID)
	require.Equal(t, "gauge", metrics[2].Type)
	require.NotNil(t, metrics[2].Value)
	require.Equal(t, baseValue+float64(j)+0.3, *metrics[2].Value)
}

func TestMakeRequest_Success(t *testing.T) {
	// Подменим client и targetURL на значения для теста
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/updates", r.URL.Path)
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var metrics []Metric
		err = json.Unmarshal(body, &metrics)
		require.NoError(t, err)

		require.Len(t, metrics, 3)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	targetURL = server.URL + "/updates"

	client = server.Client()

	metrics := initMetrics(1.0, 1, 0)
	err := makeRequest(metrics, 1, 1)
	require.NoError(t, err)
}

func TestRun(t *testing.T) {
	// создаём тестовый сервер, который всегда отвечает 200 OK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/updates" {
			t.Errorf("unexpected path: got %s, want /updates", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: got %s, want POST", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// заменяем глобальные переменные на тестовые
	client = server.Client()
	targetURL = server.URL + "/updates"

	// уменьшаем задержку запуска и количество запросов для ускорения теста
	origSleep := sleepBeforeStartSeconds
	defer func() { sleepBeforeStartSeconds = origSleep }()
	sleepBeforeStartSeconds = 0

	origConcurrency := concurrency
	origRequestsPerWorker := requestsPerWorker
	defer func() {
		concurrency = origConcurrency
		requestsPerWorker = origRequestsPerWorker
	}()
	concurrency = 2
	requestsPerWorker = 5

	run()
}
