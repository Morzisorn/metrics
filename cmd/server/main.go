package main

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/morzisorn/metrics/config"
	server "github.com/morzisorn/metrics/internal/server/handlers"
	"github.com/morzisorn/metrics/internal/server/logger"
	"github.com/morzisorn/metrics/internal/server/services/metrics"
	"github.com/morzisorn/metrics/internal/server/storage/file"
)

var (
	service *config.Service
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
	logger.Log.Info("Starting server on ", zap.String("address", service.Config.Addr))

	return mux.Run(service.Config.Addr)
}

func main() {
	service = config.GetService("server")

	if service.Config.DBConnStr == "" {
		file, err := file.NewFileStorage(service.Config.FileStoragePath)
		if err != nil {
			logger.Log.Panic("Error file loading storage", zap.Error(err))
		}
		defer file.Close()

		if service.Config.Restore {
			if err := metrics.LoadMetricsFromFile(); err != nil {
				logger.Log.Panic("Error loading metrics", zap.Error(err))
			}
		}
	}

	mux := createServer()
	if !(service.Config.StoreInterval != 0 && service.Config.DBConnStr == "") {
		if err := runServer(mux); err != nil {
			logger.Log.Panic("Error running server", zap.Error(err))
		}
	} else {
		go func(mux *gin.Engine) {
			if err := runServer(mux); err != nil {
				logger.Log.Panic("Error running server", zap.Error(err))
			}
		}(mux)

		if err := metrics.SaveMetrics(); err != nil {
			logger.Log.Panic("Error saving metrics", zap.Error(err))
		}
	}
}
