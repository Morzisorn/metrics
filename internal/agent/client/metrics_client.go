package agent

import (
	"encoding/json"
	"fmt"
	"net/http"

	agent "github.com/morzisorn/metrics/internal/agent/services"
)

type MetricsClient interface {
	SendMetric(mType string, name string, value float64) error
}

func (c *HTTPClient) SendMetric(m agent.Metric) error {
	url := fmt.Sprintf("http://%s/update/", c.BaseURL)

	body, err := json.Marshal(m)
	if err != nil {
		return err
	}

	resp, err := c.Client.R().
		SetBody(body).
		SetHeader("Content-Type", "application/json").
		Post(url)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode())
	}

	return nil
}

func (c *HTTPClient) SendMetricsByOne(m *agent.Metrics) error {
	for name, metric := range m.Metrics {
		err := c.SendMetric(metric)
		if err != nil {
			fmt.Println(err)
		}
		if name == agent.CounterMetric {
			*m.Metrics[agent.CounterMetric].Delta = 0
		}
	}

	return nil
}

func (c *HTTPClient) SendMetricsBatch(m *agent.Metrics) error {
	url := fmt.Sprintf("http://%s/updates/", c.BaseURL)

	slice := make([]agent.Metric, len(m.Metrics))
	var i int
	for _, metric := range m.Metrics {
		slice[i] = metric
		i++
	}

	body, err := json.Marshal(slice)
	if err != nil {
		return err
	}

	resp, err := c.Client.R().
		SetBody(body).
		SetHeader("Content-Type", "application/json").
		Post(url)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode())
	}

	*m.Metrics[agent.CounterMetric].Delta = 0

	return nil
}
