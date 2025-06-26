package agent

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/morzisorn/metrics/config"
	agent "github.com/morzisorn/metrics/internal/agent/services"
)

// SendMetric sends single metric to server
func (c *HTTPClient) SendMetric(m *agent.Metric) error {
	url := fmt.Sprintf("http://%s/update/", c.BaseURL)

	body, err := json.Marshal(m)
	if err != nil {
		return err
	}

	resp, err := c.Client.R().
		SetBody(body).
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Real-IP", config.GetService().Config.Addr).
		Post(url)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode())
	}

	return nil
}

// SendMetricsBatch sends all slice of metrics in one request
func (c *HTTPClient) SendMetricsBatch(m *agent.Metrics) error {
	url := fmt.Sprintf("http://%s/updates/", c.BaseURL)

	slice := make([]agent.Metric, len(m.Metrics))
	var i int
	m.Mu.RLock()
	for _, metric := range m.Metrics {
		slice[i] = metric
		i++
	}
	m.Mu.RUnlock()

	body, err := json.Marshal(slice)
	if err != nil {
		return err
	}

	resp, err := c.Client.R().
		SetBody(body).
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Real-IP", config.GetService().Config.Addr).
		Post(url)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode())
	}

	m.ResetCounter()

	return nil
}

func (c *HTTPClient) Close() {
	c.Client.Close()
}