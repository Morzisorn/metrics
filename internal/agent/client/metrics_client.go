package agent

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	agent "github.com/morzisorn/metrics/internal/agent/services"
)

type MetricsClient interface {
	SendMetric(mType string, name string, value float64) error
}

type HTTPClient struct {
	BaseURL string
	Client  *resty.Client
}

func (c *HTTPClient) SendMetric(m agent.Metric) error {
	url := fmt.Sprintf("http://%s/update/", c.BaseURL)

	resp, err := c.Client.R().
		SetBody(m).
		Post(url)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode())
	}

	return nil
}

func (c *HTTPClient) SendMetrics(m *agent.Metrics) error {
	for name, metric := range m.Metrics {
		err := c.SendMetric(metric)
		if err != nil {
			fmt.Println(err)
		}
		if name == agent.CounterMetric {
			var zero int64 = 0
			metric.Delta = &zero
			m.Metrics[agent.CounterMetric] = metric
		}
	}

	return nil
}
