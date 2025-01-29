package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPollAllMetrics(t *testing.T) {
	m := Metrics{}
	require.NoError(t, m.PollMetrics())
	assert.NotEmpty(t, m.RandomValue)
	assert.Equal(t, int64(1), m.PollCount)
}

func TestSendAllMetrics(t *testing.T) {
	m := Metrics{}
	require.NoError(t, m.PollMetrics())
	require.NoError(t, m.SendMetrics())
}
