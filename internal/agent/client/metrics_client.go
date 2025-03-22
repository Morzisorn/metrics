package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

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

func (c *HTTPClient) metricSenderJob(chIn chan agent.Metric, wg *sync.WaitGroup) {
	defer wg.Done()
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

	var wg sync.WaitGroup

	rateLimit := config.GetService().Config.RateLimit
	c.runWorkers(chIn, &wg, int(rateLimit))

	m.LoadMetricsToChan(chIn)

	wg.Wait()

	m.ResetCounter()

	return nil
}

func (c *HTTPClient) runWorkers(chIn chan agent.Metric, wg *sync.WaitGroup, rateLimit int) {
	for w := 0; w < rateLimit; w++ {
		wg.Add(1)
		go c.metricSenderJob(chIn, wg)
	}
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

	m.ResetCounter()

	return nil
}
