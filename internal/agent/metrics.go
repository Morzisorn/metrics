package agent

import (
	"fmt"
	"math"
	"net/http"
	"reflect"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
)

const host = "http://localhost:8080"

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

func SendMetric(client *resty.Client, mType string, gauge string, value float64) error {
	var url string
	switch mType {
	case "counter":
		url = fmt.Sprintf("%s/update/%s/%s/%d", host, "counter", gauge, int64(value))
	case "gauge":
		url = fmt.Sprintf("%s/update/%s/%s/%f", host, "gauge", gauge, value)
	default:
		return fmt.Errorf("unsupported metric type %s", mType)
	}
	resp, err := client.R().Post(url)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode())
	}

	return nil
}

func (m *Metrics) SendMetrics(client *resty.Client) error {
	for gauge, value := range m.RuntimeGauges {
		err := SendMetric(client, "gauge", gauge, value)
		if err != nil {
			fmt.Println(err)
		}
	}
	err := SendMetric(client, "gauge", "RandomValue", m.RandomValue)
	if err != nil {
		fmt.Println(err)
	}
	err = SendMetric(client, "counter", "PollCount", float64(m.PollCount))
	if err != nil {
		fmt.Println(err)
	}
	m.PollCount = 0
	return nil
}
