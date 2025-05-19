package agent

import (
	"sync"
	"testing"

	"github.com/morzisorn/metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPollAllMetrics(t *testing.T) {
	m := Metrics{}
	require.NoError(t, m.PollMetrics())
	assert.NotEmpty(t, m.Metrics)

}

func BenchmarkPollAllMetrics(b *testing.B) {
	tries := 10000
	m := Metrics{}

	for i := 0; i < tries; i++ {
		err := m.PollMetrics()
		if err != nil {
			b.Fatal("")
		}
	}
}

/*
func TestGetMetric(t *testing.T) {
	var memStats runtime.MemStats
	memStats.Alloc = 123456
	memStats.GCCPUFraction = 0.42
	memStats.NumGC = 99

	val := reflect.ValueOf(memStats)

	tests := []struct {
		name     string
		gauge    string
		expected float64
		wantErr  bool
	}{
		{"Alloc", "Alloc", 123456.0, false},
		{"GCCPUFraction", "GCCPUFraction", 0.42, false},
		{"NumGC", "NumGC", 99.0, false},
		{"InvalidField", "NonExistent", 0, true}, // Проверка для несуществующего поля
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMetric(val, tt.gauge)

			if tt.wantErr {
				assert.Error(t, err, fmt.Sprintf("ожидалась ошибка для %s", tt.gauge))
			} else {
				assert.NoError(t, err, fmt.Sprintf("ошибки быть не должно для %s", tt.gauge))
				assert.Equal(t, tt.expected, *got.Value, fmt.Sprintf("значение должно быть %v для %s", tt.expected, tt.gauge))
			}
		})
	}
}
*/

func TestLoadMetricsToChan(t *testing.T) {
	delta := int64(7)
	value := 5.3
	m := Metrics{
		Metrics: map[string]Metric{
			"PollCount": {
				Metric: models.Metric{
					ID:    "PollCount",
					MType: "counter",
					Delta: &delta,
				},
			},
			"GaugeMetric": {
				Metric: models.Metric{
					ID:    "GaugeMetric",
					MType: "gauge",
					Value: &value,
				},
			},
		},
	}

	ch := make(chan Metric)
	defer close(ch)

	go m.LoadMetricsToChan(ch)

	counter := <-ch
	gauge := <-ch

	assert.Equal(t, *m.Metrics["PollCount"].Delta, *counter.Delta)
	assert.Equal(t, *m.Metrics["GaugeMetric"].Value, *gauge.Value)
}

func TestCollectMemCPU(t *testing.T) {
	var wg sync.WaitGroup

	m := newEmptyMetrics()

	wg.Add(1)
	m.collectMemCPU(&wg)

	wg.Wait()
	assert.NotNil(t, m.Metrics["CPUutilization1"].Value)
	assert.NotNil(t, m.Metrics["TotalMemory"].MType)
	assert.NotNil(t, m.Metrics["FreeMemory"].Value)
}

func TestSetRandom(t *testing.T) {
	m := newEmptyMetrics()
	m.setRandom()

	assert.NotNil(t, m.Metrics[RandomValueMetric].Value)
}

func TestSetCounter(t *testing.T) {
	m := newEmptyMetrics()
	c := int64(2)
	m.setCounter(&c)

	assert.Equal(t, int64(2), *m.Metrics[CounterMetric].Delta)
}

func newEmptyMetrics() Metrics {
	return Metrics{
		Metrics: map[string]Metric{},
	}
}
