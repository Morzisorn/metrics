package metrics

import (
	"fmt"
	"strings"

	"github.com/morzisorn/metrics/internal/server/storage"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (m *Metrics) GetMetric() error {
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

func GetMetrics() map[string]string {
	s := storage.GetStorage()
	metrics := s.GetMetrics()
	var metricsTrimmed = make(map[string]string)
	for key, value := range metrics {
		metricsTrimmed[key] = trimTrailingZeros(fmt.Sprintf("%f", value))
	}
	return metricsTrimmed
}

func (m *Metrics) UpdateMetric() error {
	switch m.MType {
	case "counter":
		s := storage.GetStorage()
		updated, err := s.UpdateCounter(m.ID, float64(*m.Delta))
		if err != nil {
			return err
		}
		*m.Delta = int64(updated)
	case "gauge":
		s := storage.GetStorage()
		/*
			if m.ID == "RandomValue" {
				value := GetRandomValue()
				m.Value = &value
			}
		*/
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
