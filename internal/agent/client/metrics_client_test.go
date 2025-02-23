package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	agent "github.com/morzisorn/metrics/internal/agent/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	metrics := &agent.Metrics{
		RuntimeGauges: map[string]float64{
			"HeapAlloc":  12345.67,
			"StackInUse": 9876.54,
		},
		RandomValue: 42.42,
		PollCount:   10,
	}

	err := client.SendMetrics(metrics)
	assert.NoError(t, err, "SendMetrics должен завершаться без ошибок")
	assert.Equal(t, int64(0), metrics.PollCount, "PollCount должен сбрасываться в 0")
}

func TestSendAllMetrics(t *testing.T) {
	m := agent.Metrics{}
	s, c := setupTestServer()
	defer s.Close()

	require.NoError(t, m.PollMetrics())
	require.NoError(t, c.SendMetrics(&m))
}
