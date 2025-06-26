package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"net/http"
	_ "net/http/pprof"

	"github.com/gin-contrib/pprof"
	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/server/controllers/grpc_ctrl"
	pb "github.com/morzisorn/metrics/internal/proto"
	"github.com/morzisorn/metrics/internal/server/controllers/rest"
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
	mc *rest.MetricController,
	pc *rest.PagesController,
	hc *rest.HealthController,
) *http.Server {
	gin.SetMode(gin.ReleaseMode)
	mux := gin.Default()
	mux.Use(
		logger.LoggerMiddleware(),
		rest.DecryptMiddleware(),
		rest.GzipMiddleware(),
		rest.SignMiddleware(),
	)

	if mc != nil {
		registerMetricsRoutes(mux, mc)
	}
	registerPagesRoutes(mux, pc)
	registerHealthRoutes(mux, hc)

	pprof.Register(mux)

	srv := &http.Server{
		Addr:    cnfg.Config.Addr,
		Handler: mux,
	}

	return srv
}

func registerMetricsRoutes(mux *gin.Engine, mc *rest.MetricController) {
	mux.POST("/update/:type/:metric/:value", mc.UpdateMetricParams)
	mux.POST("/update/", mc.UpdateMetricBody)
	mux.POST("/updates/", mc.UpdateMetrics)
	mux.GET("/value/:type/:metric", mc.GetMetricParams)
	mux.POST("/value/", mc.GetMetricBody)
}

func registerPagesRoutes(mux *gin.Engine, pc *rest.PagesController) {
	mux.GET("/", pc.GetMetricsPage)
}

func registerHealthRoutes(mux *gin.Engine, hc *rest.HealthController) {
	mux.GET("/ping", hc.PingDB)
}

func runHTTPServer(srv *http.Server, storage *repositories.Storage) {
	logger.Log.Info("Starting server on ", zap.String("address", cnfg.Config.Addr))
	idleConnsClosed := make(chan struct{})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		shutdown(ctx, srv, idleConnsClosed, storage)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Log.Fatal("listen: %s\n", zap.Error(err))
	}

	<-idleConnsClosed

	logger.Log.Info("Server shutted down gracefully")
}

func shutdown(ctx context.Context, srv *http.Server, idleConnsClosed chan struct{}, storage *repositories.Storage) {
	logger.Log.Info("Shutdown server")

	(*storage).Close()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal("shutdown server error: %s\n", zap.Error(err))
	}
	close(idleConnsClosed)
}

func main() {
	if err := logger.Init(); err != nil {
		logger.Log.Fatal("Init logger error: ", zap.Error(err))
	}

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

	switch cnfg.Config.Protocol {
	case "http":
		createAnRunHTTPServer(metricsService, pagesService, healthService, &storage)
	case "grpc":
		createAnRunHTTPServer(nil, pagesService, healthService, &storage)
		createAndRunGRPCServer(metricsService)
	default:
		createAnRunHTTPServer(metricsService, pagesService, healthService, &storage)
	}

}

func createAnRunHTTPServer(ms *metrics.MetricService, ps *pages.PagesService, hs *health.HealthService, storage *repositories.Storage) {
	var metricsController *rest.MetricController
	if ms != nil {
		metricsController = rest.NewMetricController(ms)
	}

	pagesController := rest.NewPagesController(ps)
	healthController := rest.NewHealthController(hs)

	srv := createServer(metricsController, pagesController, healthController)

	if cnfg.Config.StoreInterval != 0 && cnfg.Config.DBConnStr == "" {
		go func() {
			if err := ms.SaveMetrics(); err != nil {
				logger.Log.Panic("Error saving metrics", zap.Error(err))
			}
		}()
	}

	runHTTPServer(srv, storage)
}

func createAndRunGRPCServer(ms *metrics.MetricService) {
	metricController := grpc_ctrl.NewMetricController(ms)
	cnfg := config.GetService()
	listen, err := net.Listen("tcp", cnfg.Config.Addr)
	if err != nil {
		logger.Log.Fatal("create listener error", zap.Error(err))
	}

	s := grpc.NewServer()
	pb.RegisterMetricControllerServer(s, metricController)

	logger.Log.Info("Run grpc server")
	
	if err := s.Serve(listen); err != nil {
		logger.Log.Fatal("Failed to run grpc server", zap.Error(err))
	}
}
