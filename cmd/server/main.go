package main

import (
	"fmt"

	"github.com/gin-gonic/gin"

	server "github.com/morzisorn/metrics/internal/server/handlers"
)

func createServer() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	mux := gin.Default()
	mux.GET("/", server.GetMetrics)
	mux.POST("/update/:type/:metric/:value", server.Update)
	mux.GET("/value/:type/:metric", server.GetMetric)
	return mux
}

func runServer(mux *gin.Engine) error {
	fmt.Println("Running server on", flagServerAddr)
	return mux.Run(flagServerAddr)
}

func main() {
	mux := createServer()
	parseServerFlags()
	if err := runServer(mux); err != nil {
		panic(err)
	}
}
