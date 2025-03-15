package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/morzisorn/metrics/internal/server/logger"
	"go.uber.org/zap"
)

var (
	retryDelays = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	retriableErrors = []string{
		pgerrcode.SerializationFailure,
		pgerrcode.DeadlockDetected,
		pgerrcode.TooManyConnections,
		pgerrcode.ObjectNotInPrerequisiteState,
		pgerrcode.QueryCanceled,
		pgerrcode.UniqueViolation,
	}
)

func PingDB(db *DBStorage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	return db.Pool.Ping(ctx)
}

func (db *DBStorage) UpdateGauge(name string, value float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second) //CHANGE TO 3
	defer cancel()

	db.mu.Lock()
	defer db.mu.Unlock()

	var val float64
	query := "INSERT INTO metrics(name, value) VALUES($1, $2) ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value RETURNING value"

	_, err := db.retryQueryRow(ctx, query, &val, name, value)
	if err != nil {
		return err
	}

	return nil
}

func (db *DBStorage) UpdateCounter(name string, value float64) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	db.mu.Lock()
	defer db.mu.Unlock()

	var val float64

	query := `INSERT INTO metrics(name, value) 
		VALUES($1, $2) 
		ON CONFLICT (name) DO UPDATE 
		SET value = metrics.value + EXCLUDED.value 
		RETURNING value`

	fl, err := db.retryQueryRow(ctx, query, &val, name, value)

	if err != nil {
		return 0, err
	}
	val = fl.(float64)

	return val, nil
}

func (db *DBStorage) UpdateCounters(metrics *map[string]float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	db.mu.Lock()
	defer db.mu.Unlock()

	query := "INSERT INTO metrics(name, value) VALUES "
	args := []interface{}{}
	placeholders := []string{}

	i := 1
	for m, v := range *metrics {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", i, i+1))
		args = append(args, m, v)
		i += 2
	}

	query += strings.Join(placeholders, ", ")

	query += " ON CONFLICT (name) DO UPDATE SET value = metrics.value + EXCLUDED.value;"

	_, err := db.retryExec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (db *DBStorage) UpdateGauges(metrics *map[string]float64) error {
	return db.WriteMetrics(metrics)
}

func (db *DBStorage) GetMetric(name string) (float64, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	db.mu.RLock()
	defer db.mu.RUnlock()

	var val float64
	query := "SELECT value FROM metrics WHERE name = $1"

	fl, err := db.retryQueryRow(ctx, query, &val, name)
	if err != nil {
		return 0, false
	}

	val = fl.(float64)

	return val, true
}

func (db *DBStorage) GetMetrics() (*map[string]float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	db.mu.RLock()
	defer db.mu.RUnlock()

	query := "SELECT name, value FROM metrics"

	rows, err := db.retryQuery(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metrics := make(map[string]float64)
	var name string
	var value float64

	for rows.Next() {
		err := rows.Scan(&name, &value)
		if err != nil {
			return nil, err
		}

		metrics[name] = value
	}

	if rows.Err() != nil {
		return nil, err
	}

	return &metrics, nil
}

func (db *DBStorage) WriteMetrics(metrics *map[string]float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	db.mu.Lock()
	defer db.mu.Unlock()

	query := "INSERT INTO metrics(name, value) VALUES "
	args := []interface{}{}
	placeholders := []string{}

	i := 1
	for m, v := range *metrics {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", i, i+1))
		args = append(args, m, v)
		i += 2
	}

	query += strings.Join(placeholders, ", ")

	query += " ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value;"

	_, err := db.retryExec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (db *DBStorage) Close() {
	db.Pool.Close()
}

func (db *DBStorage) retryQueryRow(ctx context.Context, query string, result interface{}, args ...interface{}) (interface{}, error) {
	var err error
	for i := 0; i < len(retryDelays); i++ {
		row := db.Pool.QueryRow(ctx, query, args...)
		err := row.Scan(&result)

		if err == nil {
			return result, nil
		}
		logger.Log.Error("db query error: ", zap.Error(err))

		if containsRetriableErr(err.Error()) {
			continue
		}
		return nil, err
	}
	return nil, err
}

func (db *DBStorage) retryQuery(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	var err error
	var rows pgx.Rows

	for i := 0; i < len(retryDelays); i++ {
		rows, err = db.Pool.Query(ctx, query, args...)

		if err == nil {
			return rows, nil
		}

		if containsRetriableErr(err.Error()) {
			continue
		}
		return nil, err
	}
	return nil, err
}

func (db *DBStorage) retryExec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	var err error

	for i := 0; i < len(retryDelays); i++ {
		tag, err := db.Pool.Exec(ctx, query, args...)

		if err == nil {
			return tag, nil
		}

		if containsRetriableErr(err.Error()) {
			continue
		}
		return pgconn.CommandTag{}, err
	}
	return pgconn.CommandTag{}, err
}

func containsRetriableErr(item string) bool {
	for _, i := range retriableErrors {
		if i == item {
			return true
		}
	}
	return false
}
