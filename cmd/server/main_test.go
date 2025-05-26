package main

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/server/controllers"
	"github.com/morzisorn/metrics/internal/server/repositories"
	"github.com/morzisorn/metrics/internal/server/services/health"
	"github.com/morzisorn/metrics/internal/server/services/metrics"
	"github.com/morzisorn/metrics/internal/server/services/pages"
	"github.com/stretchr/testify/require"
)

func TestCreateServer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cnfg = config.GetService("server")

	storage := repositories.NewStorage(cnfg.Config)
	metricsService := metrics.NewMetricService(storage)
	pagesService := pages.NewPagesService(metricsService)
	healthService := health.NewHealthService(storage)

	metricsController := controllers.NewMetricController(metricsService)
	pagesController := controllers.NewPagesController(pagesService)
	healthController := controllers.NewHealthController(healthService)

	router := createServer(metricsController, pagesController, healthController)

	require.NotNil(t, router)
}
