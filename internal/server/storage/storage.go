package storage

import (
	"strconv"
	"sync"
)

type Storage interface {
	GetMetric(name string) (float64, bool)
	GetMetrics() map[string]float64
	UpdateCounter(name, value string) error
	UpdateGauge(name, value string) error
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

func (m *MemStorage) UpdateCounter(name, value string) error {
	v, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return err
	}
	metric, _ := m.GetMetric(name)
	metric += float64(v)
	m.Metrics[name] = metric
	return nil
}

func (m *MemStorage) UpdateGauge(name, value string) error {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return err
	}
	m.Metrics[name] = v

	return nil
}

func (m *MemStorage) Reset() {
	m.Metrics = make(map[string]float64)
}
