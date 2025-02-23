package metrics

import (
	"testing"

	"github.com/morzisorn/metrics/internal/server/storage"
	"github.com/stretchr/testify/assert"
)

func TestGetMetric(t *testing.T) {
	s := storage.GetStorage()

	// Metric exists
	expectedValue := "42.123000"
	s.UpdateGauge("test_metric", expectedValue)

	result, err := GetMetric("test_metric")

	assert.NoError(t, err, "Expected no error for existing metric")
	assert.Equal(t, "42.123", result, "Expected trimmed metric value")

	// Metric does not exist
	result, err = GetMetric("non_existent_metric")

	assert.Error(t, err, "Expected error for non-existent metric")
	assert.Equal(t, "", result, "Expected empty string for missing metric")
}

func TestGetMetrics(t *testing.T) {
	s := storage.GetStorage()
	s.Reset()

	// Adding test metrics
	s.UpdateGauge("metric1", "10.500000")
	s.UpdateGauge("metric2", "20.0")
	s.UpdateGauge("metric3", "30.123456")

	metrics := GetMetrics()

	expected := map[string]string{
		"metric1": "10.5",
		"metric2": "20",
		"metric3": "30.123456",
	}

	assert.Equal(t, expected, metrics, "Expected correctly trimmed metric values")
}

func TestUpdateMetric(t *testing.T) {
	// Test updating a counter metric
	err := UpdateMetric("counter", "counter_metric", "5")
	assert.NoError(t, err, "Expected no error for updating counter metric")
	s := storage.GetStorage()
	have, _ := s.GetMetric("counter_metric")
	assert.Equal(t, 5.0, have, "Expected updated counter metric value")

	// Test updating a gauge metric
	err = UpdateMetric("gauge", "gauge_metric", "15.678")
	assert.NoError(t, err, "Expected no error for updating gauge metric")

	// Test invalid metric type
	err = UpdateMetric("invalid_type", "metric_invalid", "123")
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
