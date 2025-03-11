package pages

import (
	"fmt"

	"github.com/morzisorn/metrics/internal/server/services/metrics"
)

type PagesService struct {
	metrics *metrics.MetricService
}

func NewPagesService(metrics *metrics.MetricService) *PagesService {
	return &PagesService{metrics: metrics}
}

func (ps *PagesService) MetricsPage() (string, error) {
	html := "<html><head><title>Metrics</title></head><body><h1>Metrics</h1><ul>"
	metrics, err := ps.metrics.GetMetricsStr()
	if err != nil {
		return "", err
	}
	for k, v := range *metrics {
		html += fmt.Sprintf("<li>%s: %v</li>", k, v)
	}
	html += "</ul></body></html>"
	return html, nil
}
