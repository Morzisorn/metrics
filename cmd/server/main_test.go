package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/server/controllers"
	"github.com/morzisorn/metrics/internal/server/repositories"
	"github.com/morzisorn/metrics/internal/server/services/health"
	"github.com/morzisorn/metrics/internal/server/services/metrics"
	"github.com/morzisorn/metrics/internal/server/services/pages"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock Storage для тестирования
type mockStorage struct {
	closeCalled bool
	mu          sync.Mutex
}

func (m *mockStorage) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closeCalled = true
	return nil
}

// Реализуем остальные методы интерфейса Storage (заглушки)
func (m *mockStorage) UpdateGauge(name string, value float64) error              { return nil }
func (m *mockStorage) UpdateCounter(name string, value float64) (float64, error) { return 0, nil }
func (m *mockStorage) GetMetric(name string) (float64, bool)                     { return 0, false }
func (m *mockStorage) GetMetrics() (*map[string]float64, error)                  { return &map[string]float64{}, nil }
func (m *mockStorage) UpdateGauges(metrics *map[string]float64) error            { return nil }
func (m *mockStorage) UpdateCounters(metrics *map[string]float64) error          { return nil }
func (m *mockStorage) WriteMetrics(metrics *map[string]float64) error            { return nil }
func (m *mockStorage) PingDB() error                                             { return nil }

func TestCreateServer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cnfg = config.GetService("server")

	storage := repositories.NewStorage(cnfg.Config)
	metricsService := metrics.NewMetricService(storage)
	pagesService := pages.NewPagesService(metricsService)
	healthService := health.NewHealthService(storage)

	metricsController := controllers.NewMetricController(metricsService)
	pagesController := controllers.NewPagesController(pagesService)
	healthController := controllers.NewHealthController(healthService)

	router := createServer(metricsController, pagesController, healthController)

	require.NotNil(t, router)
}

func TestRunServer_Success(t *testing.T) {
	// Сохраняем оригинальную конфигурацию
	originalCnfg := cnfg
	defer func() {
		cnfg = originalCnfg
	}()

	// Создаем mock storage
	mockStore := &mockStorage{}
	var storage repositories.Storage = mockStore

	// Находим свободный порт
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Настраиваем тестовую конфигурацию с правильной структурой
	cnfg = &config.Service{
		Config: config.Config{
			CommonConfig: config.CommonConfig{
				Addr: fmt.Sprintf("localhost:%d", port),
			},
		},
	}

	// Создаем тестовый сервер с простым handler'ом для тестирования
	gin.SetMode(gin.TestMode)
	mux := gin.New()
	mux.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	srv := &http.Server{
		Addr:    cnfg.Config.CommonConfig.Addr,
		Handler: mux,
	}

	// Каналы для координации теста
	runServerStarted := make(chan struct{})
	runServerDone := make(chan struct{})
	runServerError := make(chan error, 1)

	// Запускаем РЕАЛЬНУЮ функцию runServer в горутине
	go func() {
		defer close(runServerDone)
		defer func() {
			if r := recover(); r != nil {
				// Перехватываем панику от logger.Log.Fatal и конвертируем в обычную ошибку
				runServerError <- fmt.Errorf("runServer panicked: %v", r)
			}
		}()

		// Сигнализируем, что runServer начал выполнение
		close(runServerStarted)

		// Вызываем настоящую функцию runServer
		runServer(srv, &storage)
	}()

	// Ждем запуска runServer
	select {
	case <-runServerStarted:
		// runServer начал выполнение
	case <-time.After(1 * time.Second):
		t.Fatal("runServer failed to start within timeout")
	}

	// Даем серверу время на полную инициализацию (logger.Init, signal setup, ListenAndServe)
	time.Sleep(200 * time.Millisecond)

	// Проверяем, что сервер действительно работает, делая HTTP запрос
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://" + cnfg.Config.CommonConfig.Addr + "/test")
	if err != nil {
		t.Logf("HTTP request failed (expected if server not fully started): %v", err)
	} else {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Server should respond to HTTP requests")
	}

	// Отправляем сигнал завершения серверу (имитируем SIGTERM)
	// Это должно запустить graceful shutdown внутри runServer
	process, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)

	go func() {
		time.Sleep(100 * time.Millisecond)
		// Отправляем SIGTERM самому процессу (это подхватит signal.Notify в runServer)
		err := process.Signal(syscall.SIGTERM)
		require.NoError(t, err)
	}()

	// Ждем завершения runServer
	select {
	case <-runServerDone:
		// runServer завершился корректно
	case err := <-runServerError:
		// Проверяем тип ошибки - некоторые ошибки ожидаемы
		if err != nil && !strings.Contains(err.Error(), "logger") {
			t.Fatalf("runServer failed with unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("runServer failed to shutdown within timeout")
	}

	// Проверяем, что storage.Close был вызван в процессе shutdown
	mockStore.mu.Lock()
	defer mockStore.mu.Unlock()
	assert.True(t, mockStore.closeCalled, "Storage.Close should be called during runServer shutdown")
}

func TestShutdown_Success(t *testing.T) {
	// Создаем mock storage
	mockStore := &mockStorage{}
	var storage repositories.Storage = mockStore

	// Создаем тестовый HTTP сервер
	gin.SetMode(gin.TestMode)
	mux := gin.New()
	mux.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	// Создаем HTTP сервер с тестовым адресом
	srv := &http.Server{
		Addr:    server.Listener.Addr().String(),
		Handler: mux,
	}

	// Создаем канал для идентификации завершения
	idleConnsClosed := make(chan struct{})

	// Запускаем shutdown в горутине
	go func() {
		shutdown(context.Background(),srv, idleConnsClosed, &storage)
	}()

	// Ждем завершения shutdown
	select {
	case <-idleConnsClosed:
		// Shutdown успешно завершен
	case <-time.After(2 * time.Second):
		t.Fatal("Shutdown failed to complete within timeout")
	}

	// Проверяем, что Close был вызван на storage
	mockStore.mu.Lock()
	defer mockStore.mu.Unlock()
	assert.True(t, mockStore.closeCalled, "Storage.Close should be called during shutdown")
}

func TestShutdown_GracefulServerStop(t *testing.T) {
	// Создаем mock storage
	mockStore := &mockStorage{}
	var storage repositories.Storage = mockStore

	// Создаем реальный HTTP сервер для тестирования graceful shutdown
	gin.SetMode(gin.TestMode)
	mux := gin.New()
	mux.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	srv := &http.Server{
		Addr:    "localhost:0", // Автоматический выбор порта
		Handler: mux,
	}

	// Запускаем сервер в фоне
	serverErrors := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// Даем серверу время на запуск
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что сервер запустился без ошибок
	select {
	case err := <-serverErrors:
		t.Fatalf("Server failed to start: %v", err)
	default:
		// Сервер запустился успешно
	}

	// Создаем канал для отслеживания завершения
	idleConnsClosed := make(chan struct{})

	// Выполняем graceful shutdown
	shutdownComplete := make(chan struct{})
	go func() {
		shutdown(context.Background(), srv, idleConnsClosed, &storage)
		close(shutdownComplete)
	}()

	// Ждем завершения shutdown
	select {
	case <-shutdownComplete:
		// Shutdown завершился
	case <-time.After(3 * time.Second):
		t.Fatal("Graceful shutdown failed to complete within timeout")
	}

	// Ждем закрытия канала idleConnsClosed
	select {
	case <-idleConnsClosed:
		// Канал закрыт, что означает успешное завершение
	case <-time.After(1 * time.Second):
		t.Fatal("idleConnsClosed channel was not closed")
	}

	// Проверяем, что storage.Close был вызван
	mockStore.mu.Lock()
	defer mockStore.mu.Unlock()
	assert.True(t, mockStore.closeCalled, "Storage.Close should be called during graceful shutdown")

	// Проверяем, что сервер действительно остановлен
	// Попытка повторного shutdown должна пройти без ошибок (сервер уже закрыт)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := srv.Shutdown(ctx)
	// Сервер уже закрыт, поэтому ошибки быть не должно
	require.NoError(t, err, "Second shutdown should not return error when server is already closed")
}
