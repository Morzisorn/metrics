package rest

import (
	"testing"

	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/server/repositories"
	"github.com/morzisorn/metrics/internal/server/services/health"
	"github.com/stretchr/testify/require"
)

func TestNewHealthController(t *testing.T) {
	cfg := config.GetService("server")
	cfg.Config.StorageType = "memory"
	storage := repositories.NewStorage(cfg.Config)
	service := health.NewHealthService(storage)
	controller := NewHealthController(service)
	require.NotNil(t, controller)
}
