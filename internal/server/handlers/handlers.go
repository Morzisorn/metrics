package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/morzisorn/metrics/internal/server/storage"
)

func GetMetrics(c *gin.Context) {
	s := storage.GetStorage()
	metrics := s.GetMetrics()

	html := "<html><head><title>Metrics</title></head><body><h1>Metrics</h1><ul>"
	for key, value := range metrics {
		html += fmt.Sprintf("<li>%s: %v</li>", key, value)
	}
	html += "</ul></body></html>"

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func Update(c *gin.Context) {
	if c.Request.Method != http.MethodPost || c.GetHeader("Content-Type") != "text/plain" {
		c.String(http.StatusMethodNotAllowed, "Invalid method or content type")
		return
	}

	nameMetric := c.Param("metric")
	if nameMetric == "" {
		c.String(http.StatusNotFound, "Invalid metric name")
		return
	}

	valueMetric := c.Param("value")
	if valueMetric == "" {
		c.String(http.StatusNotFound, "Invalid metric value")
		return
	}

	typeMetric := c.Param("type")
	switch typeMetric {
	case "counter":
		s := storage.GetStorage()
		err := s.UpdateCounter(nameMetric, valueMetric)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid metric value")
			return
		}
	case "gauge":
		s := storage.GetStorage()
		err := s.UpdateGauge(nameMetric, valueMetric)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid value metric")
			return
		}
	default:
		c.String(http.StatusBadRequest, "Invalid metric type")
		return
	}
	c.String(http.StatusOK, "OK")
}

func GetMetric(c *gin.Context) {
	nameMetric := c.Params.ByName("metric")
	if nameMetric == "" {
		c.String(http.StatusNotFound, "Invalid metric name")
		return
	}
	s := storage.GetStorage()
	m, exist := s.GetMetric(nameMetric)
	if !exist {
		c.String(http.StatusNotFound, "Metric not found")
		return
	}
	c.String(http.StatusOK, "%f", m)
}
