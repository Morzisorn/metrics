package file

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestXxx(t *testing.T) {
	storage, err := NewStorage("./test_file")
	require.NoError(t, err)
	require.NotNil(t, storage)

	metrics := initMetrics()

	// Test Write all metrics
	err = storage.WriteMetrics(metrics)
	require.NoError(t, err)

	// Test Get all metrics
	have, err := storage.GetMetrics()
	require.NoError(t, err)
	assert.Equal(t, *metrics, *have)

	// Test Write  and Get gauge
	err = storage.UpdateGauge("Gauge", 66.6)
	require.NoError(t, err)

	haveFloat, exist := storage.GetMetric("Gauge")
	assert.True(t, exist)
	assert.Equal(t, 66.6, haveFloat)

	// Test Write  and Get counter
	_, err = storage.UpdateCounter("Counter", 6.0)
	require.NoError(t, err)
	haveCounter, err := storage.UpdateCounter("Counter", 2.0)
	require.NoError(t, err)
	assert.Equal(t, 8.0, haveCounter)

	haveCounter, exist = storage.GetMetric("Counter")
	assert.True(t, exist)
	assert.Equal(t, 8.0, haveCounter)

	// Test update counters
	counters := map[string]float64{
		"Counter1": 3.0,
		"Counter2": 4.0,
	}
	err = storage.UpdateCounters(&counters)
	require.NoError(t, err)

	haveCounter1, exist := storage.GetMetric("Counter1")
	assert.True(t, exist)
	assert.Equal(t, 3.0, haveCounter1)

	haveCounter2, exist := storage.GetMetric("Counter2")
	assert.True(t, exist)
	assert.Equal(t, 4.0, haveCounter2)

	// Test update gauges
	gauges := map[string]float64{
		"Gauge1": 3.0,
		"Gauge2": 4.0,
	}
	err = storage.UpdateGauges(&gauges)
	require.NoError(t, err)

	haveGauge1, exist := storage.GetMetric("Gauge1")
	assert.True(t, exist)
	assert.Equal(t, 3.0, haveGauge1)

	haveGauge2, exist := storage.GetMetric("Gauge2")
	assert.True(t, exist)
	assert.Equal(t, 4.0, haveGauge2)

	storage.Close()
}
