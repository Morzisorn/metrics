package file

import (
	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/server/storage/memory"
)

func (f *FileStorage) GetMetric(name string) (float64, bool) {
	mem := memory.GetMemStorage()
	return mem.GetMetric(name)
}

func (f *FileStorage) GetMetrics() (*map[string]float64, error) {
	return f.Consumer.ReadMetrics()
}

func (f *FileStorage) UpdateCounter(name string, value float64) (float64, error) {
	mem := memory.GetMemStorage()

	var err error
	value, err = mem.UpdateCounter(name, value)
	if err != nil {
		return 0, err
	}

	if shouldWriteSync() {
		err = f.WriteMetrics(&mem.Metrics)
		if err != nil {
			return 0, err
		}
	}
	return value, nil
}

func (f *FileStorage) UpdateGauge(name string, value float64) error {
	mem := memory.GetMemStorage()

	err := mem.UpdateGauge(name, value)
	if err != nil {
		return err
	}

	if shouldWriteSync() {
		err = f.WriteMetrics(&mem.Metrics)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *FileStorage) WriteMetrics(metrics *map[string]float64) error {
	return f.Producer.WriteMetrics(metrics)
}

func (f *FileStorage) UpdateCounters(metrics *map[string]float64) error {
	mem := memory.GetMemStorage()
	err := mem.UpdateCounters(metrics)
	if err != nil {
		return err
	}

	if shouldWriteSync() {
		if err = f.WriteMetrics(&mem.Metrics); err != nil {
			return err
		}
	}

	return nil
}
func (f *FileStorage) UpdateGauges(metrics *map[string]float64) error {
	mem := memory.GetMemStorage()

	err := mem.UpdateGauges(metrics)
	if err != nil {
		return err
	}

	if shouldWriteSync() {
		if err = f.WriteMetrics(&mem.Metrics); err != nil {
			return err
		}
	}

	return nil
}

func shouldWriteSync() bool {
	service := config.GetService("server")
	return service.Config.StoreInterval == 0
}
