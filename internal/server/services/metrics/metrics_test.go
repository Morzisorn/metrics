package metrics

import (
	"testing"

	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/models"
	"github.com/morzisorn/metrics/internal/server/repositories"
	"github.com/morzisorn/metrics/internal/server/repositories/file"
	"github.com/morzisorn/metrics/internal/server/repositories/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMetric(t *testing.T) {
	cfg := config.GetService("server")
	storage := repositories.NewStorage(cfg.Config)
	service := NewMetricService(storage)

	f1 := 75.4
	i1 := int64(4)

	tests := []Metric{
		{
			models.Metric{
				ID:    "test_metric1",
				MType: "gauge",
				Value: &f1,
			},
		},
		{
			models.Metric{
				ID:    "test_metric2",
				MType: "counter",
				Delta: &i1,
			},
		},
	}

	// Update all metrics
	err := service.UpdateMetrics(&tests)
	require.NoError(t, err)

	// Get 1 metric
	find := Metric{
		models.Metric{
			ID:    "test_metric1",
			MType: "gauge",
			Value: new(float64),
		},
	}
	err = service.GetMetric(&find)

	assert.NoError(t, err)
	assert.Equal(t, f1, *find.Value)

	// Get unknown
	unknown := Metric{
		models.Metric{
			ID:    "unknown_metric",
			MType: "gauge",
		},
	}
	err = service.GetMetric(&unknown)
	require.Error(t, err)
}

func TestGetMetricsStr(t *testing.T) {
	storage := memory.GetStorage()
	storage.Reset()
	service := NewMetricService(storage)

	v1 := 75.400
	v2 := 4.3

	tests := []Metric{
		{
			models.Metric{
				ID:    "test_metric1",
				MType: "gauge",
				Value: &v1,
			},
		},
		{
			models.Metric{
				ID:    "test_metric2",
				MType: "gauge",
				Value: &v2,
			},
		},
	}

	// Update all metrics
	err := service.UpdateMetrics(&tests)
	require.NoError(t, err)

	metrics, err := service.GetMetricsStr()
	require.NoError(t, err)

	expected := map[string]string{
		"test_metric1": "75.4",
		"test_metric2": "4.3",
	}

	assert.Equal(t, expected, *metrics, "Expected correctly trimmed metric values")
}

func TestUpdateMetric(t *testing.T) {
	storage := memory.GetStorage()
	storage.Reset()
	service := NewMetricService(storage)

	f1 := 75.400
	i1 := int64(2)

	tests := []Metric{
		{
			models.Metric{
				ID:    "test_gauge",
				MType: "gauge",
				Value: &f1,
			},
		},
		{
			models.Metric{
				ID:    "test_counter",
				MType: "counter",
				Delta: &i1,
			},
		},
		{
			models.Metric{
				MType: "invalid_type",
				ID:    "metric_invalid",
				Delta: &i1,
			},
		},
	}

	// Test updating a counter metric
	findCounter := Metric{
		models.Metric{
			ID:    "test_counter",
			MType: "counter",
			Value: new(float64),
		},
	}
	err := service.UpdateMetric(&(tests[1]))
	assert.NoError(t, err)
	err = service.GetMetric(&findCounter)

	assert.Equal(t, int64(2), *findCounter.Delta)

	// Test updating a gauge metric
	findGauge := Metric{
		models.Metric{
			ID:    "test_gauge",
			MType: "gauge",
			Value: new(float64),
		},
	}

	err = service.UpdateMetric(&(tests[0]))
	assert.NoError(t, err)
	err = service.GetMetric(&findGauge)

	assert.Equal(t, 75.4, *findGauge.Value)

	// Test invalid metric type
	err = service.UpdateMetric(&(tests[2]))
	require.Error(t, err)
}

func TestLoadMetricsFromFile(t *testing.T) {
	storage1, err := file.NewStorage("./test_file")
	require.NoError(t, err)
	service1 := NewMetricService(storage1)


	v1 := 75.400
	v2 := 4.3

	tests := []Metric{
		{
			models.Metric{
				ID:    "test_metric1",
				MType: "gauge",
				Value: &v1,
			},
		},
		{
			models.Metric{
				ID:    "test_metric2",
				MType: "gauge",
				Value: &v2,
			},
		},
	}

	err = service1.UpdateMetrics(&tests)
	require.NoError(t, err)

	storage2, err := file.NewStorage("./test_file")
	require.NoError(t, err)

	service2 := NewMetricService(storage2)

	err = service2.LoadMetricsFromFile()

	require.NoError(t, err)
}

func TestTrimTrailingZeros(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"10.500000", "10.5"},
		{"20.000000", "20"},
		{"30.123000", "30.123"},
		{"40.000", "40"},
		{"50.1", "50.1"},
		{"60.", "60"},
	}

	for _, test := range tests {
		result := trimTrailingZeros(test.input)
		assert.Equal(t, test.expected, result, "Expected trimmed string")
	}
}
