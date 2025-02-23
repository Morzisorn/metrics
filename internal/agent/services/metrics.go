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
	RuntimeGauges map[string]float64

	PollCount   int64
	RandomValue float64
}

func (m *Metrics) PollMetrics() error {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.RuntimeGauges = make(map[string]float64)

	val := reflect.ValueOf(memStats)
	for _, gauge := range RuntimeGauges {
		value, err := GetMetric(val, gauge)
		if err != nil {
			return err
		}
		m.RuntimeGauges[gauge] = value
	}
	m.RandomValue = GetRandomValue()
	m.PollCount++
	return nil
}

func GetMetric(memStats reflect.Value, gauge string) (float64, error) {
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
			return 0, fmt.Errorf("unsupported type %s", field.Kind())
		}
		return value, nil
	} else {
		return 0, fmt.Errorf("unsupported type %s", field.Kind())
	}
}

func GetRandomValue() float64 {
	return math.Round(float64(time.Now().Nanosecond()) / 1000000)
}
