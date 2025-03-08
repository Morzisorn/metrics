package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/morzisorn/metrics/internal/server/logger"
	"github.com/morzisorn/metrics/internal/server/services/metrics"
	"github.com/morzisorn/metrics/internal/server/storage"
	"github.com/morzisorn/metrics/internal/server/storage/memory"
	"go.uber.org/zap"
)

const ContentTypeJSON = "application/json"

func RegisterMetricsRoutes(mux *gin.Engine) {
	mux.GET("/", GetMetrics)
	mux.POST("/update/:type/:metric/:value", UpdateMetricParams)
	mux.POST("/update/", UpdateMetricBody)
	mux.POST("/updates/", UpdateMetrics)
	mux.GET("/value/:type/:metric", GetMetricParams)
	mux.POST("/value/", GetMetricBody)

	mux.GET("/ping", PingDB)
}

func GetMetrics(c *gin.Context) {
	html := "<html><head><title>Metrics</title></head><body><h1>Metrics</h1><ul>"
	metrics, err := metrics.GetMetricsStr()
	if err != nil {
		c.Status(http.StatusInternalServerError)
	}
	for k, v := range *metrics {
		html += fmt.Sprintf("<li>%s: %v</li>", k, v)
	}
	html += "</ul></body></html>"

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func UpdateMetricParams(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.String(http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	if c.Request.Header.Get("Content-Type") != "text/plain" && c.Request.Header.Get("Content-Type") != "" {
		c.String(http.StatusMethodNotAllowed, "Invalid content type")
		return
	}
	var metric metrics.Metric
	metric.ID = c.Param("metric")
	if metric.ID == "" {
		c.String(http.StatusNotFound, "Invalid metric name")
		return
	}

	value := c.Param("value")
	if value == "" {
		c.String(http.StatusNotFound, "Invalid metric value")
		return
	}

	metric.MType = c.Param("type")

	switch metric.MType {
	case "counter":
		delta, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		metric.Delta = &delta
	case "gauge":
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		metric.Value = &val
	default:
		c.String(http.StatusBadRequest, "Invalid metric type")
		return
	}

	err := metric.UpdateMetric()
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.String(http.StatusOK, "OK")
}

func UpdateMetricBody(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.String(http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	if c.Request.Header.Get("Content-Type") != ContentTypeJSON {
		c.String(http.StatusMethodNotAllowed, "Invalid content type")
		return
	}

	var metric metrics.Metric
	if err := c.BindJSON(&metric); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	if metric.ID == "" {
		c.String(http.StatusNotFound, "Invalid metric ID")
		return
	}

	if metric.Delta == nil && metric.Value == nil {
		c.String(http.StatusNotFound, "Invalid metric value")
		return
	}

	err := metric.UpdateMetric()
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, metric)
}

func UpdateMetrics(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.String(http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	if c.Request.Header.Get("Content-Type") != ContentTypeJSON {
		c.String(http.StatusMethodNotAllowed, "Invalid content type")
		return
	}

	var met []metrics.Metric

	if err := c.BindJSON(&met); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	err := metrics.UpdateMetrics(&met)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	updated := memory.GetMemStorage().Metrics

	slice := make([]metrics.Metric, len(updated))
	var i int

	for name, value := range updated {
		var metric metrics.Metric
		if name == "PollCount" {
			var v = int64(value)
			metric.ID = name
			metric.MType = "counter"
			metric.Delta = &v
		} else {
			metric.ID = name
			metric.MType = "gauge"
			metric.Value = &value
		}
		slice[i] = metric
		i++
	}

	c.JSON(http.StatusOK, slice)
}

func GetMetricParams(c *gin.Context) {
	var metric metrics.Metric
	metric.ID = c.Params.ByName("metric")
	if metric.ID == "" {
		c.String(http.StatusNotFound, "Invalid metric name")
		return
	}

	metric.MType = c.Params.ByName("type")

	err := metric.GetMetric()
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
		return
	}

	switch metric.MType {
	case "counter":
		c.JSON(http.StatusOK, metric.Delta)
	case "gauge":
		c.JSON(http.StatusOK, metric.Value)
	default:
		c.String(http.StatusBadRequest, "Invalid metric type")
	}
}

func GetMetricBody(c *gin.Context) {
	if c.Request.Header.Get("Content-Type") != ContentTypeJSON {
		logger.Log.Info("Invalid content type", zap.String("Content-Type :", c.Request.Header.Get("Content-Type")))
		c.String(http.StatusMethodNotAllowed, "Invalid content type")
		return
	}

	var metric metrics.Metric
	if err := c.BindJSON(&metric); err != nil {
		logger.Log.Info("Invalid request body", zap.Error(err))
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	if metric.ID == "" || metric.MType == "" {
		logger.Log.Info("Invalid metric ID or Mtype", zap.String("ID", metric.ID), zap.String("MType", metric.MType))
		c.String(http.StatusNotFound, "Invalid metric ID or Mtype")
		return
	}

	err := metric.GetMetric()
	if err != nil {
		metrics, _ := storage.GetStorage().GetMetrics()
		logger.Log.Info("Metric not found", zap.Error(err))
		fmt.Println("Mem storage: ", metrics)
		c.String(http.StatusNotFound, err.Error())
		return
	}

	c.JSON(http.StatusOK, metric)
}
