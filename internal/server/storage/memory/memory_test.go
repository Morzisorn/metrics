package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestMemStorage_Reset(t *testing.T) {
	s := GetMemStorage()

	_, err := s.UpdateCounter("test_metric", 1)
	assert.NoError(t, err)

	_, exist := s.GetMetric("test_metric")
	assert.True(t, exist)

	s.Reset()

	_, exist = s.GetMetric("test_metric")
	assert.False(t, exist)
}
