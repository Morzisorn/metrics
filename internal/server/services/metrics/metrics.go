package metrics

import (
	"fmt"
	"strings"

	"github.com/morzisorn/metrics/internal/server/storage"
)

func GetMetric(name string) (string, error) {
	s := storage.GetStorage()
	m, exist := s.GetMetric(name)
	if !exist {
		return "", fmt.Errorf("metric not found")
	}

	return trimTrailingZeros(fmt.Sprintf("%f", m)), nil
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

func UpdateMetric(typ, name, value string) error {
	switch typ {
	case "counter":
		s := storage.GetStorage()
		err := s.UpdateCounter(name, value)
		if err != nil {
			return err
		}
	case "gauge":
		s := storage.GetStorage()
		err := s.UpdateGauge(name, value)
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
