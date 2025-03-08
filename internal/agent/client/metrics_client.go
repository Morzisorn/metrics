package agent

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/morzisorn/metrics/config"
	agent "github.com/morzisorn/metrics/internal/agent/services"
	"resty.dev/v3"
)

type MetricsClient interface {
	SendMetric(mType string, name string, value float64) error
}

type HTTPClient struct {
	BaseURL string
	Client  *resty.Client
}

func NewClient(s *config.Service) *HTTPClient {
	c := HTTPClient{
		BaseURL: s.Config.Addr,
		Client:  resty.New().SetBaseURL(s.Config.Addr),
	}

	c.Client.AddRequestMiddleware(func(client *resty.Client, req *resty.Request) error {
		err := gzipMiddleware(req)
		if err != nil {
			return err
		}

		return nil
	})

	return &c
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
	for _, metric := range m.Metrics {
		err := c.SendMetric(metric)
		if err != nil {
			fmt.Println(err)
		}
		if metric.ID == agent.CounterMetric {
			*metric.Delta = 0
		}
	}

	return nil
}

func (c *HTTPClient) SendMetricsBatch(m *agent.Metrics) error {
	url := fmt.Sprintf("http://%s/updates/", c.BaseURL)
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
