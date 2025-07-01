package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/server/controllers/restapi"
	"github.com/morzisorn/metrics/internal/server/logger"
	"github.com/morzisorn/metrics/internal/server/repositories"
	"github.com/morzisorn/metrics/internal/server/services/health"
	"github.com/morzisorn/metrics/internal/server/services/metrics"
	"github.com/morzisorn/metrics/internal/server/services/pages"
	"go.uber.org/zap"
)

type HTTPServer struct {
	Server  *http.Server
	Storage *repositories.Storage
	MS      *metrics.MetricService
}

func createHTTPServer(
	ms *metrics.MetricService,
	ps *pages.PagesService,
	hs *health.HealthService,
	storage *repositories.Storage,
) *HTTPServer {
	var mc *restapi.MetricController
	if ms != nil {
		mc = restapi.NewMetricController(ms)
	}

	pc := restapi.NewPagesController(ps)
	hc := restapi.NewHealthController(hs)

	gin.SetMode(gin.ReleaseMode)
	mux := gin.Default()
	mux.Use(
		logger.LoggerMiddleware(),
		restapi.DecryptMiddleware(),
		restapi.GzipMiddleware(),
		restapi.SignMiddleware(),
	)

	if mc != nil {
		registerMetricsRoutes(mux, mc)
	}
	registerPagesRoutes(mux, pc)
	registerHealthRoutes(mux, hc)

	pprof.Register(mux)

	srv := &http.Server{
		Addr:    config.GetService().Config.Addr,
		Handler: mux,
	}

	return &HTTPServer{srv, storage, ms}
}

func registerMetricsRoutes(mux *gin.Engine, mc *restapi.MetricController) {
	mux.POST("/update/:type/:metric/:value", mc.UpdateMetricParams)
	mux.POST("/update/", mc.UpdateMetricBody)
	mux.POST("/updates/", mc.UpdateMetrics)
	mux.GET("/value/:type/:metric", mc.GetMetricParams)
	mux.POST("/value/", mc.GetMetricBody)
}

func registerPagesRoutes(mux *gin.Engine, pc *restapi.PagesController) {
	mux.GET("/", pc.GetMetricsPage)
}

func registerHealthRoutes(mux *gin.Engine, hc *restapi.HealthController) {
	mux.GET("/ping", hc.PingDB)
}

func (s *HTTPServer) Run() {
	cnfg := config.GetService()

	logger.Log.Info("Starting server on ", zap.String("address", cnfg.Config.Addr))
	idleConnsClosed := make(chan struct{})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		s.Shutdown(ctx, idleConnsClosed)
	}()

	if cnfg.Config.StoreInterval != 0 && cnfg.Config.DBConnStr == "" {
		go func() {
			if err := s.MS.SaveMetrics(); err != nil {
				logger.Log.Panic("Error saving metrics", zap.Error(err))
			}
		}()
	}

	if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Log.Fatal("listen: %s\n", zap.Error(err))
	}

	<-idleConnsClosed

	logger.Log.Info("Server shutted down gracefully")
}

func (s *HTTPServer) Shutdown(ctx context.Context, idleConnsClosed chan struct{}) {
	logger.Log.Info("Shutdown server")

	(*s.Storage).Close()

	if err := s.Server.Shutdown(ctx); err != nil {
		logger.Log.Fatal("shutdown server error: %s\n", zap.Error(err))
	}
	close(idleConnsClosed)
}
