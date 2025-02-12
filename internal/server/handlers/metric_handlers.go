package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/morzisorn/metrics/internal/server/services/metrics"
)

func RegisterMetricsRoutes(mux *gin.Engine) {
	mux.GET("/", GetMetrics)
	mux.POST("/update/:type/:metric/:value", UpdateMetrics)
	mux.GET("/value/:type/:metric", GetMetric)
}

func GetMetrics(c *gin.Context) {

	html := "<html><head><title>Metrics</title></head><body><h1>Metrics</h1><ul>"
	for k, v := range metrics.GetMetrics() {
		html += fmt.Sprintf("<li>%s: %v</li>", k, v)
	}
	html += "</ul></body></html>"

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func UpdateMetrics(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.String(http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	if c.Request.Header.Get("Content-Type") != "text/plain" && c.Request.Header.Get("Content-Type") != "" {
		c.String(http.StatusMethodNotAllowed, "Invalid content type")
		return
	}

	name := c.Param("metric")
	if name == "" {
		c.String(http.StatusNotFound, "Invalid metric name")
		return
	}

	value := c.Param("value")
	if value == "" {
		c.String(http.StatusNotFound, "Invalid metric value")
		return
	}

	typ := c.Param("type")

	err := metrics.UpdateMetric(typ, name, value)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.String(http.StatusOK, "OK")
}

func GetMetric(c *gin.Context) {
	name := c.Params.ByName("metric")
	if name == "" {
		c.String(http.StatusNotFound, "Invalid metric name")
		return
	}
	value, err := metrics.GetMetric(name)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
		return
	}
	c.String(http.StatusOK, value)
}
