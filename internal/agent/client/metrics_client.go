package agent

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/morzisorn/metrics/config"
	agent "github.com/morzisorn/metrics/internal/agent/services"
	"github.com/morzisorn/metrics/internal/server/logger"
	"go.uber.org/zap"
)

type MetricsClient interface {
	SendMetric(mType string, name string, value float64) error
}

func (c *HTTPClient) SendMetric(m *agent.Metric) error {
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

func (c *HTTPClient) sender(chIn chan agent.Metric) {
	for metric := range chIn {
		err := c.SendMetric(&metric)
		if err != nil {
			logger.Log.Error("Send metric error. ",
				zap.String("Metric: ", metric.ID),
				zap.Error(err),
			)
			return
		}
	}
}

func (c *HTTPClient) SendMetricsByOne(m *agent.Metrics) error {
	chIn := make(chan agent.Metric, len(m.Metrics))
	defer close(chIn)

	rateLimit := config.GetService("agent").Config.RateLimit

	for w := 0; w < int(rateLimit); w++ {
		go c.sender(chIn)
	}

	m.Mu.RLock()
	for _, metric := range m.Metrics {
		chIn <- metric
	}
	m.Mu.RUnlock()

	m.Mu.Lock()
	if m.Metrics[agent.CounterMetric].Delta != nil {
		*m.Metrics[agent.CounterMetric].Delta = 0
	}
	m.Mu.Unlock()

	return nil
}

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
		Post(url)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode())
	}

	m.Mu.Lock()
	if m.Metrics[agent.CounterMetric].Delta != nil {
		*m.Metrics[agent.CounterMetric].Delta = 0
	}
	m.Mu.Unlock()

	return nil
}
