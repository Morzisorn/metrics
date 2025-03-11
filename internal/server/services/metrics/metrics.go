package metrics

import (
	"fmt"
	"strings"
	"time"

	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/models"
	rep "github.com/morzisorn/metrics/internal/server/repositories"
	"github.com/morzisorn/metrics/internal/server/repositories/memory"
)

type MetricService struct {
	storage rep.Storage
}

type Metric struct {
	models.Metric
}

func NewMetricService(storage rep.Storage) *MetricService {
	return &MetricService{storage: storage}
}

func (ms *MetricService) GetMetric(m *Metric) error {
	val, exist := ms.storage.GetMetric(m.ID)
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

func (ms *MetricService) GetMetricsStr() (*map[string]string, error) {
	metrics, err := ms.storage.GetMetrics()
	if err != nil {
		return nil, err
	}

	var metricsTrimmed = make(map[string]string)
	for key, value := range *metrics {
		metricsTrimmed[key] = trimTrailingZeros(fmt.Sprintf("%f", value))
	}
	return &metricsTrimmed, nil
}

func (ms *MetricService) UpdateMetric(m *Metric) error {
	switch m.MType {
	case "counter":
		updated, err := ms.storage.UpdateCounter(m.ID, float64(*m.Delta))
		if err != nil {
			return err
		}
		*m.Delta = int64(updated)
	case "gauge":
		err := ms.storage.UpdateGauge(m.ID, *m.Value)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid metric type")
	}

	return nil
}

func (ms *MetricService) UpdateMetrics(metrics *[]Metric) error {
	var gauges = make(map[string]float64)
	var counters = make(map[string]float64)

	for _, m := range *metrics {
		err := m.CheckMetric()
		if err != nil {
			return err
		}

		switch m.MType {
		case "counter":
			counters[m.ID] += float64(*m.Delta)

		case "gauge":
			gauges[m.ID] = *m.Value
		}
	}

	if len(counters) > 0 {
		err := ms.storage.UpdateCounters(&counters)
		if err != nil {
			return err
		}
	}

	if len(gauges) > 0 {
		err := ms.storage.UpdateGauges(&gauges)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Metric) CheckMetric() error {
	if m.ID == "" {
		return fmt.Errorf("invalid metric ID")
	}

	if m.Delta == nil && m.Value == nil {
		return fmt.Errorf("invalid metric value")
	}
	return nil
}

func trimTrailingZeros(s string) string {
	s = strings.TrimRight(s, "0")
	s = strings.TrimSuffix(s, ".")

	return s
}

func (ms *MetricService) LoadMetricsFromFile() error {
	service := config.GetService("server")
	mem := memory.GetStorage()

	if service.Config.Restore {
		metrics, err := ms.storage.GetMetrics()
		if err != nil {
			return err
		}
		err = mem.WriteMetrics(metrics)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ms *MetricService) SaveMetrics() error {
	lastSave := time.Now()
	service := config.GetService("server")
	mem := memory.GetStorage()

	for {
		if time.Since(lastSave).Seconds() >= float64(service.Config.StoreInterval) {
			lastSave = time.Now()

			metrics, err := mem.GetMetrics()
			if err != nil {
				return err
			}

			err = ms.storage.WriteMetrics(metrics)
			if err != nil {
				return err
			}
		}
	}
}
