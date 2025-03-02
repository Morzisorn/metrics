package main

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/morzisorn/metrics/config"
	server "github.com/morzisorn/metrics/internal/server/handlers"
	"github.com/morzisorn/metrics/internal/server/logger"
	"github.com/morzisorn/metrics/internal/server/services/metrics"
	"github.com/morzisorn/metrics/internal/server/storage"
)

var (
	Service *config.Service
	File    *storage.FileStorage
)

func createServer() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	mux := gin.Default()
	mux.Use(
		logger.LoggerMiddleware(),
		server.GzipMiddleware(),
	)

	server.RegisterMetricsRoutes(mux)

	return mux
}

func runServer(mux *gin.Engine) error {
	if err := logger.Init(); err != nil {
		return err
	}
	logger.Log.Info("Starting server on ", zap.String("address", Service.Config.Addr))

	return mux.Run(Service.Config.Addr)
}

func main() {
	var err error
	Service = config.GetService("server")

	File, err = storage.NewFileStorage(Service.Config.FileStoragePath)
	if err != nil {
		logger.Log.Panic("Error file loading storage", zap.Error(err))
	}
	defer File.Close()

	if Service.Config.Restore {
		if err := metrics.LoadMetricsFromFile(); err != nil {
			logger.Log.Panic("Error loading metrics", zap.Error(err))
		}
	}

	mux := createServer()
	go func(mux *gin.Engine) {
		if err := runServer(mux); err != nil {
			logger.Log.Panic("Error running server", zap.Error(err))
		}
	}(mux)

	if Service.Config.StoreInterval != 0 {
		if err := metrics.SaveMetrics(); err != nil {
			logger.Log.Panic("Error saving metrics", zap.Error(err))
		}
	}
}
