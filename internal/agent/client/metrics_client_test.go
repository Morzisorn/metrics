package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"

	agent "github.com/morzisorn/metrics/internal/agent/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"resty.dev/v3"
)

func setupTestServer() (*httptest.Server, *HTTPClient) {
	handler := http.NewServeMux()
	handler.HandleFunc("/update/counter/PollCount/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler.HandleFunc("/update/gauge/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)

	client := &HTTPClient{
		BaseURL: server.URL[len("http://"):],
		Client:  resty.New(),
	}

	return server, client
}

func TestSendMetrics(t *testing.T) {
	server, client := setupTestServer()
	defer server.Close()

	randomValue := 42.42
	metrics := &agent.Metrics{
		Metrics: map[string]agent.Metric{
			"HeapAlloc": {
				ID:    "HeapAlloc",
				MType: "gauge",
				Value: &randomValue,
			},
		},
	}

	err := client.SendMetrics(metrics)
	assert.NoError(t, err, "SendMetrics must not return an error")
	assert.Equal(t, randomValue, *metrics.Metrics["HeapAlloc"].Value, "Incorrect value after sending metrics")
}

func TestSendAllMetrics(t *testing.T) {
	m := agent.Metrics{}
	s, c := setupTestServer()
	defer s.Close()

	require.NoError(t, m.PollMetrics())
	require.NoError(t, c.SendMetrics(&m))
}
