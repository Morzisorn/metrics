package storage

import (
	"sync"
)

type Storage interface {
	GetMetric(name string) (float64, bool)
	GetMetrics() map[string]float64
	UpdateCounter(name string, value float64) (float64, error)
	UpdateGauge(name string, value float64) error
	SetMetrics(metrics map[string]float64)
	Reset()
}

type MemStorage struct {
	Metrics map[string]float64
}

var (
	instance Storage
	once     sync.Once
)

func GetStorage() Storage {
	once.Do(func() {
		instance = &MemStorage{
			Metrics: make(map[string]float64),
		}
		instance.(*MemStorage).Metrics["RandomValue"] = 5.6
	})
	return instance
}

func (m *MemStorage) GetMetric(name string) (float64, bool) {
	metric, exist := m.Metrics[name]
	return metric, exist
}

func (m *MemStorage) GetMetrics() map[string]float64 {
	return m.Metrics
}

func (m *MemStorage) UpdateCounter(name string, value float64) (float64, error) {
	metric, _ := m.GetMetric(name)
	metric += float64(value)
	m.Metrics[name] = metric
	return metric, nil
}

func (m *MemStorage) UpdateGauge(name string, value float64) error {
	m.Metrics[name] = value

	return nil
}

func (m *MemStorage) SetMetrics(metrics map[string]float64) {
	m.Metrics = metrics
}

func (m *MemStorage) Reset() {
	m.Metrics = make(map[string]float64)
}
