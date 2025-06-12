package pages

import (
	"testing"

	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/models"
	"github.com/morzisorn/metrics/internal/server/repositories"
	"github.com/morzisorn/metrics/internal/server/services/metrics"
	"github.com/stretchr/testify/assert"
)

func generateMetricService() *metrics.MetricService {
	cnfg := config.GetService("server").Config
	storage := repositories.NewStorage(cnfg)
	return metrics.NewMetricService(storage)
}

func TestNewHealthService(t *testing.T) {
	metricsServ := generateMetricService()
	pagesServ := NewPagesService(metricsServ)
	assert.NotNil(t, pagesServ)
}

func TestMetricsPage(t *testing.T) {
	ms := generateMetricService()
	value := 7.4
	metric := metrics.Metric{
		Metric: models.Metric{
			ID:    "TestMetric",
			MType: "gauge",
			Value: &value,
		},
	}
	err := ms.UpdateMetric(&metric)
	assert.NoError(t, err)

	ps := NewPagesService(ms)
	str, err := ps.MetricsPage()
	assert.NotEqual(t, "", str)
	assert.NoError(t, err)
}
