package file

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initMetrics() *map[string]float64 {
	return &map[string]float64{
		"TestMetric1": 265.47,
		"TestMetric2": 23.09,
	}
}

func TestNewStorage(t *testing.T) {
	storage, err := NewStorage("./test_file")
	require.NoError(t, err)
	require.NotNil(t, storage)

	metrics := initMetrics()

	err = storage.WriteMetrics(metrics)
	require.NoError(t, err)

	have, err := storage.GetMetrics()
	require.NoError(t, err)

	assert.Equal(t, *metrics, *have)

	storage.Close()
}
