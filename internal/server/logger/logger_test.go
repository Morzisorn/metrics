package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	err := Init()
	require.NoError(t, err)
}

func TestLoggerMiddleware(t *testing.T) {
	// Инициализируем логгер
	err := Init()
	require.NoError(t, err)

	// Включаем test mode в gin
	gin.SetMode(gin.TestMode)

	// Создаем тестовый роутер с middleware
	router := gin.New()
	router.Use(LoggerMiddleware())

	// Маршрут без ошибки
	router.GET("/success", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// Маршрут с ошибкой
	router.GET("/fail", func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "fail"})
	})

	// Тестируем успешный запрос
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/success", nil)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// Тестируем ошибочный запрос
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/fail", nil)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusInternalServerError, w.Code)
}
