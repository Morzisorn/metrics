package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/morzisorn/metrics/internal/server/storage"
)

func GetMetrics(c *gin.Context) {
	s := storage.GetStorage()
	metrics := s.GetMetrics()

	html := "<html><head><title>Metrics</title></head><body><h1>Metrics</h1><ul>"
	for key, value := range metrics {
		str := trimTrailingZeros(fmt.Sprintf("%f", value))
		html += fmt.Sprintf("<li>%s: %v</li>", key, str)
	}
	html += "</ul></body></html>"

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func Update(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.String(http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	if c.Request.Header.Get("Content-Type") != "text/plain" && c.Request.Header.Get("Content-Type") != "" {
		c.String(http.StatusMethodNotAllowed, "Invalid content type")
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
	str := trimTrailingZeros(fmt.Sprintf("%f", m))
	c.String(http.StatusOK, str)
}

func trimTrailingZeros(s string) string {
	s = strings.TrimRight(s, "0") // Убираем нули справа
	s = strings.TrimSuffix(s, ".")

	return s
}
