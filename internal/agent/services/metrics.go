package agent

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"time"
)

var RuntimeGauges = []string{
	"Alloc",
	"BuckHashSys",
	"Frees",
	"GCCPUFraction",
	"GCSys",
	"HeapAlloc",
	"HeapIdle",
	"HeapInuse",
	"HeapObjects",
	"HeapReleased",
	"HeapSys",
	"LastGC",
	"Lookups",
	"MCacheInuse",
	"MCacheSys",
	"MSpanInuse",
	"MSpanSys",
	"Mallocs",
	"NextGC",
	"NumForcedGC",
	"NumGC",
	"OtherSys",
	"PauseTotalNs",
	"StackInuse",
	"StackSys",
	"Sys",
	"TotalAlloc",
}

type MetricsCollector interface {
	PollMetrics() error
}

type Metrics struct {
	Metrics map[string]Metric
}

const (
	CounterMetric     = "PollCount"   //int64
	RandomValueMetric = "RandomValue" //float64
)

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (m *Metrics) PollMetrics() error {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.Metrics = make(map[string]Metric)

	val := reflect.ValueOf(memStats)
	for _, gauge := range RuntimeGauges {
		metr, err := GetMetric(val, gauge)
		if err != nil {
			return err
		}
		m.Metrics[gauge] = metr
	}

	m.Metrics[RandomValueMetric] = Metric{
		ID:    RandomValueMetric,
		MType: "gauge",
		Value: GetRandomValue(),
	}

	var counter int64 = 1

	m.Metrics[CounterMetric] = Metric{
		ID:    CounterMetric,
		MType: "counter",
		Delta: &counter,
	}

	return nil
}

func GetMetric(memStats reflect.Value, gauge string) (Metric, error) {
	field := memStats.FieldByName(gauge)
	if field.IsValid() {
		var value float64
		switch field.Kind() {
		case reflect.Uint64:
			value = float64(field.Uint())
		case reflect.Uint32:
			value = float64(field.Uint())
		case reflect.Float64:
			value = field.Float()
		default:
			return Metric{}, fmt.Errorf("unsupported type %s", field.Kind())
		}
		return Metric{
			ID:    gauge,
			MType: "gauge",
			Value: &value,
		}, nil
	} else {
		return Metric{}, fmt.Errorf("unsupported type %s", field.Kind())
	}
}

func GetRandomValue() *float64 {
	v := math.Round(float64(time.Now().Nanosecond()) / 1000000)
	return &v
}
