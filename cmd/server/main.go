package main

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/server/controllers"
	"github.com/morzisorn/metrics/internal/server/logger"
	"github.com/morzisorn/metrics/internal/server/repositories"
	"github.com/morzisorn/metrics/internal/server/services/health"
	"github.com/morzisorn/metrics/internal/server/services/metrics"
	"github.com/morzisorn/metrics/internal/server/services/pages"
)

var (
	cnfg *config.Service
)

func createServer(
	mc *controllers.MetricController,
	pc *controllers.PagesController,
	hc *controllers.HealthController,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	mux := gin.Default()
	mux.Use(
		logger.LoggerMiddleware(),
		controllers.GzipMiddleware(),
		controllers.SignMiddleware(),
	)

	registerMetricsRoutes(mux, mc)
	registerPagesRoutes(mux, pc)
	registerHealthRoutes(mux, hc)

	return mux
}

func registerMetricsRoutes(mux *gin.Engine, mc *controllers.MetricController) {
	mux.POST("/update/:type/:metric/:value", mc.UpdateMetricParams)
	mux.POST("/update/", mc.UpdateMetricBody)
	mux.POST("/updates/", mc.UpdateMetrics)
	mux.GET("/value/:type/:metric", mc.GetMetricParams)
	mux.POST("/value/", mc.GetMetricBody)
}

func registerPagesRoutes(mux *gin.Engine, pc *controllers.PagesController) {
	mux.GET("/", pc.GetMetricsPage)
}

func registerHealthRoutes(mux *gin.Engine, hc *controllers.HealthController) {
	mux.GET("/ping", hc.PingDB)
}

func runServer(mux *gin.Engine) error {
	if err := logger.Init(); err != nil {
		return err
	}
	logger.Log.Info("Starting server on ", zap.String("address", cnfg.Config.Addr))

	return mux.Run(cnfg.Config.Addr)
}

func main() {
	cnfg = config.GetService("server")

	storage := repositories.NewStorage(cnfg.Config)
	metricsService := metrics.NewMetricService(storage)
	pagesService := pages.NewPagesService(metricsService)
	healthService := health.NewHealthService(storage)

	if cnfg.Config.StorageType == "file" && cnfg.Config.Restore {
		if err := metricsService.LoadMetricsFromFile(); err != nil {
			logger.Log.Panic("Error loading metrics", zap.Error(err))
		}
	}

	metricsController := controllers.NewMetricController(metricsService)
	pagesController := controllers.NewPagesController(pagesService)
	healthController := controllers.NewHealthController(healthService)

	mux := createServer(metricsController, pagesController, healthController)

	if cnfg.Config.StoreInterval != 0 && cnfg.Config.DBConnStr == "" {
		go func() {
			if err := metricsService.SaveMetrics(); err != nil {
				logger.Log.Panic("Error saving metrics", zap.Error(err))
			}
		}()
	}

	if err := runServer(mux); err != nil {
		logger.Log.Panic("Error running server", zap.Error(err))
	}
}
