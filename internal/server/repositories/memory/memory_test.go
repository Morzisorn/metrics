package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMetric(t *testing.T) {
	m := MemStorage{Metrics: map[string]float64{"test": 1}}
	v, exist := m.GetMetric("test")
	assert.True(t, exist)
	assert.Equal(t, 1.0, v)
}

func TestUpdateCounter(t *testing.T) {
	m := MemStorage{Metrics: map[string]float64{"test": 1}}
	updated, err := m.UpdateCounter("test", 1)
	assert.NoError(t, err)
	assert.Equal(t, 2.0, updated)
}

func TestUpdateGauge(t *testing.T) {
	m := MemStorage{Metrics: map[string]float64{"test": 1}}
	err := m.UpdateGauge("test", 2)
	assert.NoError(t, err)
	v, _ := m.GetMetric("test")
	assert.Equal(t, 2.0, v)
}

func TestWriteAndGetMetrics(t *testing.T) {
	s := GetStorage()
	s.Reset()

	metrics := map[string]float64{
		"Test1": 1.0,
		"Test2": 2.5,
	}

	err := s.WriteMetrics(&metrics)
	require.NoError(t, err)

	haveMetrics, err := s.GetMetrics()
	require.NoError(t, err)

	assert.Equal(t, metrics["Test2"], (*haveMetrics)["Test2"])
}

func TestUpdateGauges(t *testing.T) {
	s := GetStorage()
	s.Reset()

	metrics := map[string]float64{
		"Test1": 1.0,
		"Test2": 2.5,
	}

	err := s.UpdateGauges(&metrics)
	require.NoError(t, err)

	m, exist := s.GetMetric("Test1")
	require.True(t, exist)
	assert.Equal(t, metrics["Test1"], m)
}

func TestUpdateCounters(t *testing.T) {
	s := GetStorage()
	s.Reset()

	metrics := map[string]float64{
		"Test1": 1.0,
		"Test2": 2.5,
	}

	err := s.UpdateCounters(&metrics)
	require.NoError(t, err)

	err = s.UpdateCounters(&metrics)
	require.NoError(t, err)

	m, exist := s.GetMetric("Test1")
	require.True(t, exist)
	assert.Equal(t, 2*metrics["Test1"], m)
}

func TestMemStorage_Reset(t *testing.T) {
	s := GetStorage()

	_, err := s.UpdateCounter("test_metric", 1)
	assert.NoError(t, err)

	_, exist := s.GetMetric("test_metric")
	assert.True(t, exist)

	s.Reset()

	_, exist = s.GetMetric("test_metric")
	assert.False(t, exist)
}
