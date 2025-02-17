package metrics

import (
	"testing"

	"github.com/morzisorn/metrics/internal/server/storage"
	"github.com/stretchr/testify/assert"
)

func TestGetMetric(t *testing.T) {
	s := storage.GetStorage()
	tests := []struct {
		metric Metrics
		expect float64
	}{
		{
			metric: Metrics{
				ID:    "test_metric",
				MType: "gauge",
				Value: new(float64),
			},
			expect: 42.123,
		},
		{
			metric: Metrics{
				ID:    "non_existent_metric",
				MType: "gauge",
				Value: new(float64),
			},
			expect: float64(0),
		},
	}

	// Metric exists
	err := s.UpdateGauge(tests[0].metric.ID, tests[0].expect)
	assert.NoError(t, err)

	err = tests[0].metric.GetMetric()

	assert.NoError(t, err, "Expected no error for existing metric")
	assert.Equal(t, tests[0].expect, *tests[0].metric.Value, "Expected trimmed metric value")

	// Metric does not exist
	err = tests[1].metric.GetMetric()

	assert.Error(t, err, "Expected error for non-existent metric")
	assert.Equal(t, tests[1].expect, *tests[1].metric.Value, "Expected 0 for missing metric")
}

func TestGetMetrics(t *testing.T) {
	s := storage.GetStorage()
	s.Reset()

	// Adding test metrics
	err := s.UpdateGauge("metric1", 10.5)
	assert.NoError(t, err)
	err = s.UpdateGauge("metric2", 20.0)
	assert.NoError(t, err)
	err = s.UpdateGauge("metric3", 30.123456)
	assert.NoError(t, err)

	metrics := GetMetrics()

	expected := map[string]string{
		"metric1": "10.5",
		"metric2": "20",
		"metric3": "30.123456",
	}

	assert.Equal(t, expected, metrics, "Expected correctly trimmed metric values")
}

func TestUpdateMetric(t *testing.T) {
	tests := []struct {
		metric Metrics
		err    string
	}{
		{
			metric: Metrics{
				MType: "counter",
				ID:    "counter_metric",
				Delta: new(int64),
			},
			err: "",
		},
		{
			metric: Metrics{
				MType: "gauge",
				ID:    "gauge_metric",
				Value: new(float64),
			},
			err: "",
		},
		{
			metric: Metrics{
				MType: "invalid_type",
				ID:    "metric_invalid",
				Delta: new(int64),
			},
			err: "invalid metric type",
		},
	}
	// Test updating a counter metric
	*tests[0].metric.Delta = 5

	err := tests[0].metric.UpdateMetric()
	assert.NoError(t, err, "Expected no error for updating counter metric")
	s := storage.GetStorage()
	have, _ := s.GetMetric(tests[0].metric.ID)
	assert.Equal(t, 5.0, have, "Expected updated counter metric value")

	// Test updating a gauge metric
	*tests[1].metric.Value = 15.678

	err = tests[1].metric.UpdateMetric()
	assert.NoError(t, err, "Expected no error for updating gauge metric")
	assert.Equal(t, 15.678, *tests[1].metric.Value, "Expected updated metric")

	// Test invalid metric type
	*tests[2].metric.Delta = 123

	err = tests[2].metric.UpdateMetric()
	assert.Error(t, err, "Expected error for invalid metric type")
	assert.Equal(t, "invalid metric type", err.Error(), "Expected specific error message")
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
