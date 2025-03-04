package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/morzisorn/metrics/internal/server/logger"
	"go.uber.org/zap"
)

func PingDB(db *pgx.Conn) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	return db.Ping(ctx)
}

func (db *DBStorage) UpdateGauge(name string, value float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var val float64
	err := db.DB.QueryRow(ctx,
		"INSERT INTO metrics(name, value) VALUES($1, $2) ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value RETURNING value",
		name, value).
		Scan(&val)

	if err != nil {
		return err
	}

	return nil
}

func (db *DBStorage) UpdateCounter(name string, value float64) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var val float64
	val, _ = db.GetMetric(name)

	value += val

	err := db.DB.QueryRow(ctx,
		"INSERT INTO metrics(name, value) VALUES($1, $2) ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value RETURNING value",
		name, value).
		Scan(&val)

	if err != nil {
		return 0, err
	}

	return val, nil
}

func (db *DBStorage) GetMetric(name string) (float64, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var val float64

	err := db.DB.QueryRow(ctx,
		"SELECT value FROM metrics WHERE name = $1", name).Scan(&val)
	if err != nil {
		return 0, false
	}

	return val, true
}

func (db *DBStorage) GetMetrics() (*map[string]float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := db.DB.Query(ctx,
		"SELECT name, value FROM metrics")
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

func (db *DBStorage) SetMetrics(metrics *map[string]float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "INSERT INTO metrics(name, value) VALUES "
	args := []interface{}{}
	placeholder := 1
	i := 1

	for m, v := range *metrics {
		query += fmt.Sprintf("(%d, %d)", placeholder, placeholder+1)
		if i < len(*metrics)-1 {
			query += ", "
		}

		args = append(args, m, v)
		placeholder += 2
		i += 1
	}
	query += "ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value;"

	_, err := db.DB.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (db *DBStorage) Close() {
	err := db.DB.Close(context.Background())
	if err != nil {
		logger.Log.Panic("DB close error: ", zap.Error(err))
	}
}
