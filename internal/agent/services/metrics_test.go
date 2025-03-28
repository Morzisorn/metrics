package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPollAllMetrics(t *testing.T) {
	m := Metrics{}
	require.NoError(t, m.PollMetrics())
	assert.NotEmpty(t, m.Metrics)

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
