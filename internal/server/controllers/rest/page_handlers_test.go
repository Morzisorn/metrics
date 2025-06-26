package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/morzisorn/metrics/internal/models"
	"github.com/morzisorn/metrics/internal/server/services/metrics"
	"github.com/morzisorn/metrics/internal/server/services/pages"
	"github.com/stretchr/testify/require"
)

func TestGetMetricsPage_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	metricService := createTestMetricService()
	metric := &metrics.Metric{
		Metric: models.Metric{
			ID:    "test5",
			MType: "counter",
			Delta: getPointer(int64(1)),
		},
	}
	err := metricService.UpdateMetric(metric)
	require.NoError(t, err)

	pagesService := pages.NewPagesService(metricService)

	controller := NewPagesController(pagesService)
	router.GET("/metrics", controller.GetMetricsPage)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
	require.Equal(t, "text/html; charset=utf-8", resp.Header().Get("Content-Type"))
	require.Contains(t, resp.Body.String(), "test5")
}

