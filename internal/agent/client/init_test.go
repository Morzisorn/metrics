package agent

import (
	"testing"

	"github.com/morzisorn/metrics/config"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	cfg := config.GetService("agent")
	client := NewClient(cfg)
	require.NotNil(t, client)
}
