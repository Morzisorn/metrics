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

func (c *HTTPClient) SendMetric(typ string, name string, value float64) error {
	var url string
	switch typ {
	case "counter":
		url = fmt.Sprintf("%s/update/%s/%s/%d", "http://"+c.BaseURL, "counter", name, int64(value))
	case "gauge":
		url = fmt.Sprintf("%s/update/%s/%s/%f", "http://"+c.BaseURL, "gauge", name, value)
	default:
		return fmt.Errorf("unsupported metric type %s", typ)
	}
	resp, err := c.Client.R().Post(url)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode())
	}

	return nil
}

func (c *HTTPClient) SendMetrics(m *agent.Metrics) error {
	for gauge, value := range m.RuntimeGauges {
		err := c.SendMetric("gauge", gauge, value)
		if err != nil {
			fmt.Println(err)
		}
	}
	err := c.SendMetric("gauge", "RandomValue", m.RandomValue)
	if err != nil {
		fmt.Println(err)
	}
	err = c.SendMetric("counter", "PollCount", float64(m.PollCount))
	if err != nil {
		fmt.Println(err)
	}
	m.PollCount = 0
	return nil
}
