package database

import (
	"errors"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/morzisorn/metrics/internal/models"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDBStorage_PingDB_Success(t *testing.T) {
	// Создаем mock pool
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	// Настраиваем ожидание для Ping
	mockPool.ExpectPing()

	// Создаем экземпляр DBStorage с mock pool
	db := &DBStorage{
		Pool: mockPool,
	}

	// Выполняем тест
	err = db.PingDB()

	// Проверяем результат
	assert.NoError(t, err)

	// Проверяем, что все ожидания были выполнены
	err = mockPool.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestDBStorage_PingDB_Error(t *testing.T) {
	// Создаем mock pool
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	// Настраиваем ожидание для Ping с ошибкой
	expectedErr := errors.New("database connection failed")
	mockPool.ExpectPing().WillReturnError(expectedErr)

	// Создаем экземпляр DBStorage с mock pool
	db := &DBStorage{
		Pool: mockPool,
	}

	// Выполняем тест
	err = db.PingDB()

	// Проверяем, что получили ожидаемую ошибку
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)

	// Проверяем, что все ожидания были выполнены
	err = mockPool.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestDBStorage_UpdateGauge_Success(t *testing.T) {
	// Создаем mock pool
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	// Создаем экземпляр DBStorage с mock pool
	db := &DBStorage{
		Pool: mockPool,
	}

	metric := models.Metric{
		ID:    "test_metric",
		Value: getPointer(4.5),
	}

	// Настраиваем ожидание для запроса INSERT ... ON CONFLICT
	expectedQuery := "INSERT INTO metrics\\(name, value\\) VALUES\\(\\$1, \\$2\\) ON CONFLICT \\(name\\) DO UPDATE SET value = EXCLUDED\\.value RETURNING value"

	// Создаем mock строку, которая будет возвращена
	rows := pgxmock.NewRows([]string{"value"}).AddRow(4.5)

	// Настраиваем ожидание запроса с конкретными параметрами
	mockPool.ExpectQuery(expectedQuery).
		WithArgs("test_metric", 4.5).
		WillReturnRows(rows)

	// Выполняем тест
	err = db.UpdateGauge(metric.ID, *metric.Value)

	// Проверяем результат
	assert.NoError(t, err)

	// Проверяем, что все ожидания были выполнены
	err = mockPool.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestDBStorage_UpdateCounter_Success(t *testing.T) {
	// Создаем mock pool
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	// Создаем экземпляр DBStorage с mock pool
	db := &DBStorage{
		Pool: mockPool,
	}

	metric := models.Metric{
		ID:    "test_metric",
		Delta: getPointer(int64(4)),
	}

	// Настраиваем ожидание для запроса INSERT ... ON CONFLICT
	// Точно такой же SQL, как в коде (без переносов)
	expectedQuery := "INSERT INTO metrics\\(name, value\\) VALUES\\(\\$1, \\$2\\) ON CONFLICT \\(name\\) DO UPDATE SET value = metrics\\.value \\+ EXCLUDED\\.value RETURNING value"

	// Создаем mock строку, которая будет возвращена
	rows := pgxmock.NewRows([]string{"value"}).AddRow(4.0)

	// Настраиваем ожидание запроса с конкретными параметрами
	mockPool.ExpectQuery(expectedQuery).
		WithArgs("test_metric", float64(4)).
		WillReturnRows(rows)

	// Выполняем тест
	res, err := db.UpdateCounter(metric.ID, float64(*metric.Delta))

	// Проверяем результат
	assert.NoError(t, err)
	assert.Equal(t, *metric.Delta, int64(res))

	// Проверяем, что все ожидания были выполнены
	err = mockPool.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestDBStorage_UpdateCounters_Success(t *testing.T) {
	// Создаем mock pool
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	// Создаем экземпляр DBStorage с mock pool
	db := &DBStorage{
		Pool: mockPool,
	}

	// Подготавливаем тестовые данные
	metrics := map[string]float64{
		"counter1": 10.0,
		"counter2": 20.0,
		"counter3": 30.0,
	}

	// Поскольку порядок итерации по map не гарантирован,
	// нужно использовать более гибкий подход для проверки SQL

	// Ожидаем SQL запрос с тремя парами значений
	// Используем регулярное выражение, которое учитывает любой порядок параметров
	expectedQuery := `INSERT INTO metrics\(name, value\) VALUES \(\$1, \$2\), \(\$3, \$4\), \(\$5, \$6\) ON CONFLICT \(name\) DO UPDATE SET value = metrics\.value \+ EXCLUDED\.value;`

	// Настраиваем ожидание для Exec (не Query, так как нет RETURNING)
	mockPool.ExpectExec(expectedQuery).
		WithArgs(
			pgxmock.AnyArg(), pgxmock.AnyArg(), // counter1 или counter2 или counter3
			pgxmock.AnyArg(), pgxmock.AnyArg(), // и их значения
			pgxmock.AnyArg(), pgxmock.AnyArg(),
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 3)) // 3 строки затронуто

	// Выполняем тест
	err = db.UpdateCounters(&metrics)

	// Проверяем результат
	assert.NoError(t, err)

	// Проверяем, что все ожидания были выполнены
	err = mockPool.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestDBStorage_UpdateGauges_Success(t *testing.T) {
	// Создаем mock pool
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	// Создаем экземпляр DBStorage с mock pool
	db := &DBStorage{
		Pool: mockPool,
	}

	// Подготавливаем тестовые данные
	metrics := map[string]float64{
		"temperature": 23.5,
		"humidity":    65.2,
		"pressure":    1013.25,
	}

	// Настраиваем ожидание для INSERT запроса с тремя парами значений
	// UpdateGauges вызывает WriteMetrics, поэтому ожидаем тот же SQL
	expectedQuery := `INSERT INTO metrics\(name, value\) VALUES \(\$1, \$2\), \(\$3, \$4\), \(\$5, \$6\) ON CONFLICT \(name\) DO UPDATE SET value = EXCLUDED\.value;`

	// Настраиваем ожидание для Exec
	mockPool.ExpectExec(expectedQuery).
		WithArgs(
			pgxmock.AnyArg(), pgxmock.AnyArg(), // любая метрика и её значение
			pgxmock.AnyArg(), pgxmock.AnyArg(), // любая метрика и её значение
			pgxmock.AnyArg(), pgxmock.AnyArg(), // любая метрика и её значение
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 3)) // 3 строки затронуто

	// Выполняем тест
	err = db.UpdateGauges(&metrics)

	// Проверяем результат
	assert.NoError(t, err)

	// Проверяем, что все ожидания были выполнены
	err = mockPool.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestDBStorage_GetMetric_Success(t *testing.T) {
	// Создаем mock pool
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	// Создаем экземпляр DBStorage с mock pool
	db := &DBStorage{
		Pool: mockPool,
	}

	// Тестовые данные
	metricName := "test_metric"
	expectedValue := 42.5

	// Настраиваем ожидание для SELECT запроса
	expectedQuery := "SELECT value FROM metrics WHERE name = \\$1"

	// Создаем mock строку с результатом
	rows := pgxmock.NewRows([]string{"value"}).AddRow(expectedValue)

	// Настраиваем ожидание запроса
	mockPool.ExpectQuery(expectedQuery).
		WithArgs(metricName).
		WillReturnRows(rows)

	// Выполняем тест
	value, found := db.GetMetric(metricName)

	// Проверяем результат
	assert.True(t, found)
	assert.Equal(t, expectedValue, value)

	// Проверяем, что все ожидания были выполнены
	err = mockPool.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestDBStorage_GetMetrics_Success(t *testing.T) {
	// Создаем mock pool
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	// Создаем экземпляр DBStorage с mock pool
	db := &DBStorage{
		Pool: mockPool,
	}

	// Настраиваем ожидание для SELECT запроса
	expectedQuery := "SELECT name, value FROM metrics"

	// Создаем mock строки с несколькими метриками
	rows := pgxmock.NewRows([]string{"name", "value"}).
		AddRow("cpu_usage", 75.5).
		AddRow("memory_usage", 60.2).
		AddRow("disk_usage", 45.8)

	// Настраиваем ожидание запроса
	mockPool.ExpectQuery(expectedQuery).
		WillReturnRows(rows)

	// Выполняем тест
	result, err := db.GetMetrics()

	// Проверяем результат
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Проверяем содержимое возвращенной карты
	expectedMetrics := map[string]float64{
		"cpu_usage":    75.5,
		"memory_usage": 60.2,
		"disk_usage":   45.8,
	}

	assert.Equal(t, expectedMetrics, *result)
	assert.Len(t, *result, 3)

	// Проверяем, что все ожидания были выполнены
	err = mockPool.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestDBStorage_WriteMetrics_Success(t *testing.T) {
	// Создаем mock pool
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	// Создаем экземпляр DBStorage с mock pool
	db := &DBStorage{
		Pool: mockPool,
	}

	// Подготавливаем тестовые данные
	metrics := map[string]float64{
		"cpu_usage":    75.5,
		"memory_usage": 60.2,
		"disk_usage":   45.8,
	}

	// Настраиваем ожидание для INSERT запроса с тремя парами значений
	// Используем регулярное выражение, которое учитывает любой порядок параметров
	expectedQuery := `INSERT INTO metrics\(name, value\) VALUES \(\$1, \$2\), \(\$3, \$4\), \(\$5, \$6\) ON CONFLICT \(name\) DO UPDATE SET value = EXCLUDED\.value;`

	// Настраиваем ожидание для Exec (не Query, так как нет RETURNING)
	mockPool.ExpectExec(expectedQuery).
		WithArgs(
			pgxmock.AnyArg(), pgxmock.AnyArg(), // любая метрика и её значение
			pgxmock.AnyArg(), pgxmock.AnyArg(), // любая метрика и её значение
			pgxmock.AnyArg(), pgxmock.AnyArg(), // любая метрика и её значение
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 3)) // 3 строки затронуто

	// Выполняем тест
	err = db.WriteMetrics(&metrics)

	// Проверяем результат
	assert.NoError(t, err)

	// Проверяем, что все ожидания были выполнены
	err = mockPool.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestDBStorage_Close_Success(t *testing.T) {
	// Создаем mock pool
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	// Создаем экземпляр DBStorage с mock pool
	db := &DBStorage{
		Pool: mockPool,
	}

	// Выполняем тест - вызываем Close
	err = db.Close()

	// Проверяем результат
	assert.NoError(t, err)

	// Примечание: pgxmock автоматически отслеживает вызов Close()
	// при defer mockPool.Close() в других тестах
	// Здесь мы проверяем, что метод Close() DBStorage не возвращает ошибку
}

func TestContainsRetriableErr_Success(t *testing.T) {
	retriableError := pgerrcode.UniqueViolation

	result := containsRetriableErr(retriableError)

	assert.True(t, result, "Expected retriable error to be found in the list")
}

func getPointer[T any](v T) *T {
	return &v
}
