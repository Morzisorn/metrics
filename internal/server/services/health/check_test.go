package health

import (
	"testing"

	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/server/repositories"
	"github.com/stretchr/testify/assert"
)

func TestNewHealthService(t *testing.T) {
	cnfg := config.GetService("server").Config
	storage := repositories.NewStorage(cnfg)
	healthServ := NewHealthService(storage)
	assert.NotNil(t, healthServ)
}

func TestPingDB(t *testing.T) {
	assert.NoError(t, PingDB())
}
