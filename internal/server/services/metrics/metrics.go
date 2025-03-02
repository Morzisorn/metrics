package metrics

import (
	"fmt"
	"strings"
	"time"

	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/models"
	"github.com/morzisorn/metrics/internal/server/storage"
)

type Metric struct {
	models.Metric
}

func (m *Metric) GetMetric() error {
	service := config.GetService("server")
	var val float64

	if service.Config.DBConnStr != "" {
		var err error
		val, err = storage.GetMetric(m.ID)
		if err != nil {
			return err
		}
	} else {
		s := storage.GetStorage()
		var exist bool
		val, exist = s.GetMetric(m.ID)
		if !exist {
			return fmt.Errorf("metric not found")
		}
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
	service := config.GetService("server")
	var metrics *map[string]float64

	if service.Config.DBConnStr != "" {
		var err error
		metrics, err = storage.GetMetrics()
		if err != nil {
			return nil, err
		}
	} else {
		s := storage.GetStorage()
		metrics = s.GetMetrics()
	}

	var metricsTrimmed = make(map[string]string)
	for key, value := range *metrics {
		metricsTrimmed[key] = trimTrailingZeros(fmt.Sprintf("%f", value))
	}
	return &metricsTrimmed, nil
}

func (m *Metric) UpdateMetric() error {
	service := config.GetService("server")

	if service.Config.DBConnStr != "" {
		switch m.MType {
		case "counter":
			updated, err := storage.WriteMetric(m.ID, float64(*m.Delta))
			if err != nil {
				return err
			}
			*m.Delta = int64(updated)
		case "gauge":
			_, err := storage.WriteMetric(m.ID, *m.Value)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid metric type")
		}
	} else {
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

		if service.Config.StoreInterval == 0 {
			file := storage.GetFileStorage()
			metrics := s.GetMetrics()
			err := file.Producer.WriteMetrics(metrics)
			if err != nil {
				return err
			}
		}
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
	file := storage.GetFileStorage()

	if service.Config.Restore {
		s := storage.GetStorage()
		metrics, err := file.Consumer.ReadMetrics()
		if err != nil {
			return err
		}
		s.SetMetrics(*metrics)
	}
	return nil
}

func SaveMetrics() error {
	lastSave := time.Now()
	service := config.GetService("server")
	file := storage.GetFileStorage()

	for {
		if time.Since(lastSave).Seconds() >= float64(service.Config.StoreInterval) {
			lastSave = time.Now()

			s := storage.GetStorage()
			metrics := s.GetMetrics()

			err := file.Producer.WriteMetrics(metrics)
			if err != nil {
				return err
			}
		}
	}
}
