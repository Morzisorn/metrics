package metrics

import (
	"fmt"
	"strings"
	"time"

	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/models"
	"github.com/morzisorn/metrics/internal/server/storage"
	"github.com/morzisorn/metrics/internal/server/storage/file"
	"github.com/morzisorn/metrics/internal/server/storage/memory"
)

type Metric struct {
	models.Metric
}

func (m *Metric) GetMetric() error {
	s := storage.GetStorage()
	val, exist := s.GetMetric(m.ID)
	if !exist {
		return fmt.Errorf("metric not found")
	}

	switch m.MType {
	case "counter":
		v := int64(val)
		m.Delta = &v
	case "gauge":
		m.Value = &val
	default:
		return fmt.Errorf("invalid metric type")
	}

	return nil
}

func GetMetricsStr() (*map[string]string, error) {
	s := storage.GetStorage()
	metrics, err := s.GetMetrics()
	if err != nil {
		return nil, err
	}

	var metricsTrimmed = make(map[string]string)
	for key, value := range *metrics {
		metricsTrimmed[key] = trimTrailingZeros(fmt.Sprintf("%f", value))
	}
	return &metricsTrimmed, nil
}

func (m *Metric) UpdateMetric() error {
	s := storage.GetStorage()

	switch m.MType {
	case "counter":
		updated, err := s.UpdateCounter(m.ID, float64(*m.Delta))
		if err != nil {
			return err
		}
		*m.Delta = int64(updated)
	case "gauge":
		err := s.UpdateGauge(m.ID, *m.Value)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid metric type")
	}

	return nil
}

func trimTrailingZeros(s string) string {
	s = strings.TrimRight(s, "0")
	s = strings.TrimSuffix(s, ".")

	return s
}

func LoadMetricsFromFile() error {
	service := config.GetService("server")
	file := file.GetFileStorage()

	if service.Config.Restore {
		metrics, err := file.Consumer.ReadMetrics()
		if err != nil {
			return err
		}
		file.SetMetrics(metrics)
	}
	return nil
}

func SaveMetrics() error {
	lastSave := time.Now()
	service := config.GetService("server")
	file := file.GetFileStorage()
	mem := memory.GetMemStorage()

	for {
		if time.Since(lastSave).Seconds() >= float64(service.Config.StoreInterval) {
			lastSave = time.Now()

			metrics, err := mem.GetMetrics()
			if err != nil {
				return err
			}

			err = file.Producer.WriteMetrics(metrics)
			if err != nil {
				return err
			}
		}
	}
}
