package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/morzisorn/metrics/config"
	server "github.com/morzisorn/metrics/internal/server/handlers"
	"github.com/morzisorn/metrics/internal/server/logger"
)

var Service *config.Service

func createServer() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	mux := gin.Default()
	mux.Use(logger.WithLogger())

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
	Service, err = config.New("server")
	if err != nil {
		panic(err)
	}
	mux := createServer()
	if err := runServer(mux); err != nil {
		fmt.Println("Server failed to start:", err)
		panic(err)
	}
}
