package agent

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/morzisorn/metrics/internal/models"
	"github.com/morzisorn/metrics/internal/server/logger"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
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
	Mu      sync.RWMutex
}

type Metric struct {
	models.Metric
}

const (
	CounterMetric     = "PollCount"   //int64
	RandomValueMetric = "RandomValue" //float64
)

func (m *Metrics) PollMetrics() error {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.Metrics = make(map[string]Metric)
	val := reflect.ValueOf(memStats)

	generator := pollGenerator(RuntimeGauges)

	m.produceMetrics(generator, &val)

	return nil
}

func pollGenerator(gauges []string) chan string {
	chIn := make(chan string)

	go func() {
		defer close(chIn)

		for _, g := range gauges {
			chIn <- g
		}
	}()

	return chIn
}

func (m *Metrics) produceMetrics(chIn chan string, refl *reflect.Value) {
	var wg sync.WaitGroup

	for s := range chIn {
		wg.Add(1)
		go func(name string) {
			metric, err := GetMetric(refl, name)
			if err != nil {
				logger.Log.Error("Get metric error: ", zap.Error(err))
			}
			m.Mu.Lock()
			m.Metrics[name] = metric
			m.Mu.Unlock()
		}(s)
		wg.Done()
	}

	wg.Add(1)
	go func() {
		var counter int64 = 1

		m.Mu.Lock()
		m.Metrics[RandomValueMetric] = Metric{
			Metric: models.Metric{
				ID:    RandomValueMetric,
				MType: "gauge",
				Value: GetRandomValue(),
			},
		}

		m.Metrics[CounterMetric] = Metric{
			Metric: models.Metric{
				ID:    CounterMetric,
				MType: "counter",
				Delta: &counter,
			},
		}
		m.Mu.Unlock()
	}()
	wg.Done()

	wg.Add(1)
	go m.collectMemCPU()
	wg.Done()

	wg.Wait()
}

func (m *Metrics) collectMemCPU() {
	percent, err := cpu.Percent(0, false)
	if err != nil {
		logger.Log.Error("Get CPUutilization1 error: ", zap.Error(err))
	}

	cpuUtilization1 := Metric{
		Metric: models.Metric{
			ID:    "CPUutilization1",
			MType: "gauge",
			Value: &percent[0],
		},
	}

	memory, err := mem.VirtualMemory()
	if err != nil {
		logger.Log.Error("Get memory error: ", zap.Error(err))
	}

	total := float64(memory.Total)
	totalMemory := Metric{
		Metric: models.Metric{
			ID:    "TotalMemory",
			MType: "gauge",
			Value: &total,
		},
	}

	free := float64(memory.Free)
	freeMemory := Metric{
		Metric: models.Metric{
			ID:    "FreeMemory",
			MType: "gauge",
			Value: &free,
		},
	}

	m.Mu.Lock()
	m.Metrics["CPUutilization1"] = cpuUtilization1
	m.Metrics["TotalMemory"] = totalMemory
	m.Metrics["FreeMemory"] = freeMemory
	m.Mu.Unlock()
}

func GetMetric(memStats *reflect.Value, gauge string) (Metric, error) {
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
			Metric: models.Metric{
				ID:    gauge,
				MType: "gauge",
				Value: &value,
			},
		}, nil
	} else {
		return Metric{}, fmt.Errorf("unsupported type %s", field.Kind())
	}
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano())) // Создаём генератор случайных чисел

func GetRandomValue() *float64 {
	v := rng.Float64()
	return &v
}
