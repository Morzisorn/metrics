package agent

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPollAllMetrics(t *testing.T) {
	m := Metrics{}
	require.NoError(t, m.PollMetrics())
	assert.NotEmpty(t, m.Metrics["HeapAlloc"])
	assert.Equal(t, int64(1), m.Metrics[CounterMetric].Delta)
}

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
		{"Alloc", "Alloc", 123456, false},
		{"GCCPUFraction", "GCCPUFraction", 0.42, false},
		{"NumGC", "NumGC", 99, false},
		{"InvalidField", "NonExistent", 0, true}, // Проверка для несуществующего поля
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMetric(val, tt.gauge)

			if tt.wantErr {
				assert.Error(t, err, fmt.Sprintf("ожидалась ошибка для %s", tt.gauge))
			} else {
				assert.NoError(t, err, fmt.Sprintf("ошибки быть не должно для %s", tt.gauge))
				assert.Equal(t, tt.expected, got, fmt.Sprintf("значение должно быть %v для %s", tt.expected, tt.gauge))
			}
		})
	}
}
