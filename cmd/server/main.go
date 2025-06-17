package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"net/http"
	_ "net/http/pprof"

	"github.com/gin-contrib/pprof"
	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/server/controllers"
	"github.com/morzisorn/metrics/internal/server/logger"
	"github.com/morzisorn/metrics/internal/server/repositories"
	"github.com/morzisorn/metrics/internal/server/services/health"
	"github.com/morzisorn/metrics/internal/server/services/metrics"
	"github.com/morzisorn/metrics/internal/server/services/pages"
)

// buildVersion represents the version of the application.
// It is injected during build using -ldflags.
// Example: go build -ldflags "-X main.buildVersion=1.0.0"
var buildVersion string

// buildDate represents the build date of the application.
// It is injected during build using -ldflags.
// Example: go build -ldflags "-X main.buildDate=2025-05-29"
var buildDate string

// buildCommit represents the commit from which the application was built.
// It is injected during build using -ldflags.
// Example: go build -ldflags "-X main.buildCommit=abc123"
var buildCommit string

var (
	cnfg *config.Service
)

func createServer(
	mc *controllers.MetricController,
	pc *controllers.PagesController,
	hc *controllers.HealthController,
) *http.Server {
	gin.SetMode(gin.ReleaseMode)
	mux := gin.Default()
	mux.Use(
		logger.LoggerMiddleware(),
		controllers.DecryptMiddleware(),
		controllers.GzipMiddleware(),
		controllers.SignMiddleware(),
	)

	registerMetricsRoutes(mux, mc)
	registerPagesRoutes(mux, pc)
	registerHealthRoutes(mux, hc)

	pprof.Register(mux)

	srv := &http.Server{
		Addr:    cnfg.Config.Addr,
		Handler: mux,
	}

	return srv
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

func runServer(srv *http.Server, storage *repositories.Storage) {
	if err := logger.Init(); err != nil {
		logger.Log.Fatal("Init logger error: ", zap.Error(err))
	}
	logger.Log.Info("Starting server on ", zap.String("address", cnfg.Config.Addr))
	idleConnsClosed := make(chan struct{})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		<-quit
		shutdown(srv, idleConnsClosed, storage)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Log.Fatal("listen: %s\n", zap.Error(err))
	}

	<-idleConnsClosed

	logger.Log.Info("Server shutted down gracefully")
}

func shutdown(srv *http.Server, idleConnsClosed chan struct{}, storage *repositories.Storage) {
	logger.Log.Info("Shutdown server")
	
	(*storage).Close()

	if err := srv.Shutdown(context.Background()); err != nil {
		logger.Log.Fatal("shutdown server error: %s\n", zap.Error(err))
	}
	close(idleConnsClosed)
}

func main() {
	config.PrintMetaInfo(buildVersion, buildDate, buildCommit)
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

	srv := createServer(metricsController, pagesController, healthController)

	if cnfg.Config.StoreInterval != 0 && cnfg.Config.DBConnStr == "" {
		go func() {
			if err := metricsService.SaveMetrics(); err != nil {
				logger.Log.Panic("Error saving metrics", zap.Error(err))
			}
		}()
	}

	runServer(srv, &storage)
}
