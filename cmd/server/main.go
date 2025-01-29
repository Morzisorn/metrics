package main

import (
	"github.com/gin-gonic/gin"

	server "github.com/morzisorn/metrics/internal/server/handlers"
)

func startServer() *gin.Engine {
	mux := gin.Default()
	mux.GET("/", server.GetMetrics)
	mux.POST("/update/:type/:metric/:value", server.Update)
	mux.GET("/value/:type/:metric", server.GetMetric)
	return mux
}

func main() {
	mux := startServer()
	mux.Run()
}
