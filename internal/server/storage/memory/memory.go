package memory

import (
	"sync"

	"github.com/morzisorn/metrics/internal/server/storage/models"
)

type MemStorage struct {
	Metrics map[string]float64
}

var (
	instanceStorage models.Storage
	onceStorage     sync.Once

	instanceMem *MemStorage
	onceMem     sync.Once
)

func GetStorage() models.Storage {
	onceStorage.Do(func() {
		instanceStorage = GetMemStorage()
	})
	return instanceStorage
}

func GetMemStorage() *MemStorage {
	onceMem.Do(func() {
		instanceMem = &MemStorage{
			Metrics: make(map[string]float64),
		}
	})

	return instanceMem
}

func (m *MemStorage) GetMetric(name string) (float64, bool) {
	metric, exist := m.Metrics[name]
	return metric, exist
}

func (m *MemStorage) GetMetrics() (*map[string]float64, error) {
	return &m.Metrics, nil
}

func (m *MemStorage) UpdateCounters(metrics *map[string]float64) error {
	for name, value := range *metrics {
		_, err := m.UpdateCounter(name, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *MemStorage) UpdateGauges(metrics *map[string]float64) error {
	for name, value := range *metrics {
		err := m.UpdateGauge(name, value)
		if err != nil {
			return err
		}
	}

	return nil
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

func (m *MemStorage) WriteMetrics(metrics *map[string]float64) error {
	m.Metrics = *metrics
	return nil
}

func (m *MemStorage) Reset() {
	m.Metrics = make(map[string]float64)
}
