package controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/models"
	"github.com/morzisorn/metrics/internal/server/repositories"
	"github.com/morzisorn/metrics/internal/server/services/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	host = "http://localhost:8080"
)

func TestUpdateCounterQueryOK(t *testing.T) {
	service := createTestMetricService()
	controller := NewMetricController(service)

	url := host + "/update/counter/test/1"
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", url, nil)
	c.Request.Header.Set("Content-Type", "text/plain")
	c.Params = gin.Params{
		{Key: "type", Value: "counter"},
		{Key: "metric", Value: "test"},
		{Key: "value", Value: "1"},
	}
	c.Request.Header.Set("Content-Type", "text/plain")

	controller.UpdateMetricParams(c)
	controller.UpdateMetricParams(c)

	assert.Equal(t, http.StatusOK, c.Writer.Status())

	m := &metrics.Metric{
		Metric: models.Metric{
			ID:    "test",
			MType: "counter",
		},
	}
	err := service.GetMetric(m)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), *m.Delta)
}

func TestUpdateGaugeQueryOK(t *testing.T) {
	service := createTestMetricService()
	controller := NewMetricController(service)

	url := host + "/update/gauge/test/2.5"
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", url, nil)
	c.Request.Header.Set("Content-Type", "text/plain")
	c.Params = gin.Params{
		{Key: "type", Value: "gauge"},
		{Key: "metric", Value: "test"},
		{Key: "value", Value: "2.5"},
	}

	controller.UpdateMetricParams(c)
	controller.UpdateMetricParams(c)

	assert.Equal(t, http.StatusOK, c.Writer.Status())

	m := &metrics.Metric{
		Metric: models.Metric{
			ID:    "test",
			MType: "gauge",
		},
	}
	err := service.GetMetric(m)
	assert.NoError(t, err)
	assert.Equal(t, 2.5, *m.Value)
}

func TestUpdateInvalidQueryPath(t *testing.T) {
	service := createTestMetricService()
	controller := NewMetricController(service)

	url := host + "/update/counter/test"
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", url, nil)
	c.Request.Header.Set("Content-Type", "text/plain")
	c.Params = gin.Params{
		{Key: "type", Value: "counter"},
		{Key: "metric", Value: "test"},
	}

	controller.UpdateMetricParams(c)

	assert.Equal(t, http.StatusNotFound, c.Writer.Status())
}

func TestUpdateInvalidQueryMethod(t *testing.T) {
	service := createTestMetricService()
	controller := NewMetricController(service)

	url := host + "/update/counter/test/1"

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", url, nil)
	c.Request.Header.Set("Content-Type", "text/plain")
	c.Params = gin.Params{
		{Key: "type", Value: "counter"},
		{Key: "metric", Value: "test"},
		{Key: "value", Value: "1"},
	}

	controller.UpdateMetricParams(c)

	assert.Equal(t, http.StatusMethodNotAllowed, c.Writer.Status())
}

func TestUpdateInvalidQueryContentType(t *testing.T) {
	service := createTestMetricService()
	controller := NewMetricController(service)

	url := host + "/update/counter/test/1"

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", url, nil)
	c.Params = gin.Params{
		{Key: "type", Value: "counter"},
		{Key: "metric", Value: "test"},
		{Key: "value", Value: "1"},
	}
	c.Request.Header.Set("Content-Type", "incorrect")

	controller.UpdateMetricParams(c)

	assert.Equal(t, http.StatusMethodNotAllowed, c.Writer.Status())
}

func TestUpdateInvalidQueryType(t *testing.T) {
	service := createTestMetricService()
	controller := NewMetricController(service)

	url := host + "/update/incorrect/test/1"

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", url, nil)
	c.Request.Header.Set("Content-Type", "text/plain")
	c.Params = gin.Params{
		{Key: "type", Value: "incorrect"},
		{Key: "metric", Value: "test"},
		{Key: "value", Value: "1"},
	}

	controller.UpdateMetricParams(c)

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
}

func TestUpdateInvalidQueryGaugeValue(t *testing.T) {
	service := createTestMetricService()
	controller := NewMetricController(service)

	url := host + "/update/gauge/test/incorrect"

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", url, nil)
	c.Request.Header.Set("Content-Type", "text/plain")
	c.Params = gin.Params{
		{Key: "type", Value: "counter"},
		{Key: "metric", Value: "test"},
		{Key: "value", Value: "incorrect"},
	}

	controller.UpdateMetricParams(c)

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
}

func TestUpdateInvalidQueryCounterValue(t *testing.T) {
	service := createTestMetricService()
	controller := NewMetricController(service)

	url := host + "/update/counter/test/2.5"

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", url, nil)
	c.Request.Header.Set("Content-Type", "text/plain")
	c.Params = gin.Params{
		{Key: "type", Value: "counter"},
		{Key: "metric", Value: "test"},
		{Key: "value", Value: "2.5"},
	}

	controller.UpdateMetricParams(c)

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
}

func TestUpdateCounterBodyOK(t *testing.T) {
	service := createTestMetricService()
	controller := NewMetricController(service)

	url := host + "/update/"
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", url, nil)
	c.Request.Header.Set("Content-Type", "application/json")

	metric := &metrics.Metric{
		Metric: models.Metric{
			ID:    "test2",
			MType: "counter",
			Delta: getPointer(int64(1)),
		},
	}
	b, err := json.Marshal(metric)
	require.NoError(t, err)

	c.Request.Body = io.NopCloser(bytes.NewReader(b))

	controller.UpdateMetricBody(c)
	//controller.UpdateMetricBody(c)

	assert.Equal(t, http.StatusOK, c.Writer.Status())

	m := &metrics.Metric{
		Metric: models.Metric{
			ID:    "test2",
			MType: "counter",
		},
	}
	err = service.GetMetric(m)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), *m.Delta)
}

func TestUpdateGaugeBodyOK(t *testing.T) {
	service := createTestMetricService()
	controller := NewMetricController(service)

	url := host + "/update/"
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", url, nil)
	c.Request.Header.Set("Content-Type", "application/json")

	metric := &metrics.Metric{
		Metric: models.Metric{
			ID:    "test",
			MType: "gauge",
			Value: getPointer(2.5),
		},
	}
	b, err := json.Marshal(metric)
	require.NoError(t, err)

	c.Request.Body = io.NopCloser(bytes.NewReader(b))

	controller.UpdateMetricBody(c)
	controller.UpdateMetricBody(c)

	assert.Equal(t, http.StatusOK, c.Writer.Status())

	m := &metrics.Metric{
		Metric: models.Metric{
			ID:    "test",
			MType: "gauge",
		},
	}
	err = service.GetMetric(m)
	assert.NoError(t, err)
	assert.Equal(t, 2.5, *m.Value)
}

func TestUpdateMetricsOK(t *testing.T) {
	service := createTestMetricService()
	controller := NewMetricController(service)

	url := host + "/updates/"
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", url, nil)
	c.Request.Header.Set("Content-Type", "application/json")

	metric := &metrics.Metric{
		Metric: models.Metric{
			ID:    "test3",
			MType: "gauge",
			Value: getPointer(2.5),
		},
	}
	b, err := json.Marshal([]metrics.Metric{*metric})
	require.NoError(t, err)

	c.Request.Body = io.NopCloser(bytes.NewReader(b))

	controller.UpdateMetrics(c)

	assert.Equal(t, http.StatusOK, c.Writer.Status())

	m := &metrics.Metric{
		Metric: models.Metric{
			ID:    "test3",
			MType: "gauge",
		},
	}
	err = service.GetMetric(m)
	assert.NoError(t, err)
	assert.Equal(t, 2.5, *m.Value)
}

func TestGetMetricParamsOK(t *testing.T) {
	service := createTestMetricService()
	controller := NewMetricController(service)

	metric := &metrics.Metric{
		Metric: models.Metric{
			ID:    "test4",
			MType: "counter",
			Delta: getPointer(int64(1)),
		},
	}
	err := service.UpdateMetric(metric)
	require.NoError(t, err)

	url := host + "/update/counter/test4"
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest("GET", url, nil)
	c.Request.Header.Set("Content-Type", "text/plain")
	c.Params = gin.Params{
		{Key: "type", Value: "counter"},
		{Key: "metric", Value: "test4"},
	}

	controller.GetMetricParams(c)

	assert.Equal(t, http.StatusOK, c.Writer.Status())
	assert.Equal(t, "1", recorder.Body.String())
}

func TestGetMetricBodyOK(t *testing.T) {
	service := createTestMetricService()
	controller := NewMetricController(service)

	metric := &metrics.Metric{
		Metric: models.Metric{
			ID:    "test5",
			MType: "counter",
			Delta: getPointer(int64(1)),
		},
	}
	err := service.UpdateMetric(metric)
	require.NoError(t, err)

	url := host + "/value/"
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest("POST", url, nil)
	c.Request.Header.Set("Content-Type", "application/json")
	b, err := json.Marshal(metric)
	require.NoError(t, err)

	c.Request.Body = io.NopCloser(bytes.NewReader(b))

	controller.GetMetricBody(c)
	var m metrics.Metric
	err = json.Unmarshal(recorder.Body.Bytes(), &m)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, c.Writer.Status())
	assert.Equal(t, int64(1), *m.Delta)
}

func createTestMetricService() *metrics.MetricService {
	cfg := config.GetService("server")
	cfg.Config.StorageType = "memory"
	storage := repositories.NewStorage(cfg.Config)
	return metrics.NewMetricService(storage)
}

func getPointer[T any](v T) *T {
	return &v
}
