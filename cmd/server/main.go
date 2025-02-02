package main

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/morzisorn/metrics/internal/config"
	server "github.com/morzisorn/metrics/internal/server/handlers"
)

var Conf config.Config

func createServer() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	mux := gin.Default()
	mux.GET("/", server.GetMetrics)
	mux.POST("/update/:type/:metric/:value", server.Update)
	mux.GET("/value/:type/:metric", server.GetMetric)
	return mux
}

func runServer(mux *gin.Engine) error {
	fmt.Println("Running server on", Conf.Addr)
	return mux.Run(Conf.Addr)
}

func main() {
	mux := createServer()
	config.ParseFlags()
	if err := runServer(mux); err != nil {
		panic(err)
	}
}
