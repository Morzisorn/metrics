package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/morzisorn/metrics/internal/server/logger"
	"github.com/morzisorn/metrics/internal/server/services/metrics"
	"go.uber.org/zap"
)

const ContentTypeJSON = "application/json"

type MetricController struct {
	service *metrics.MetricService
}

func NewMetricController(service *metrics.MetricService) *MetricController {
	return &MetricController{service: service}
}

func (mc *MetricController) UpdateMetricParams(c *gin.Context) {
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

	err := mc.service.UpdateMetric(&metric)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.String(http.StatusOK, "OK")
}

func (mc *MetricController) UpdateMetricBody(c *gin.Context) {
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

	err := mc.service.UpdateMetric(&metric)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, metric)
}

func (mc *MetricController) UpdateMetrics(c *gin.Context) {
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

	err := mc.service.UpdateMetrics(&met)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	c.Status(http.StatusOK)
}

func (mc *MetricController) GetMetricParams(c *gin.Context) {
	var metric metrics.Metric
	metric.ID = c.Params.ByName("metric")
	if metric.ID == "" {
		c.String(http.StatusNotFound, "Invalid metric name")
		return
	}

	metric.MType = c.Params.ByName("type")

	err := mc.service.GetMetric(&metric)
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

func (mc *MetricController) GetMetricBody(c *gin.Context) {
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

	err := mc.service.GetMetric(&metric)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
		return
	}

	c.JSON(http.StatusOK, metric)
}
