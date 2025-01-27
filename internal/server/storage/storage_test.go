package storage

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
	err := m.UpdateCounter("test", "1")
	assert.NoError(t, err)
	v, _ := m.GetMetric("test")
	assert.Equal(t, 2.0, v)
}

func TestUpdateGauge(t *testing.T) {
	m := MemStorage{Metrics: map[string]float64{"test": 1}}
	err := m.UpdateGauge("test", "2")
	assert.NoError(t, err)
	v, _ := m.GetMetric("test")
	assert.Equal(t, 2.0, v)
}
