package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/morzisorn/metrics/internal/server/database"
)

func PingDB() error {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	return db.Ping(ctx)
}

func WriteMetric(name string, value float64) (float64, error) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var val float64
	err := db.QueryRow(ctx,
		"INSERT INTO metrics(name, value) VALUES($1, $2) RETURNING value", name, value).Scan(&val)

	if err != nil {
		return 0, err
	}

	return val, nil
}

func GetMetric(name string) (float64, error) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var val float64

	err := db.QueryRow(ctx,
		"SELECT value FROM metrics WHERE name = $1", name).Scan(&val)
	if err != nil {
		return 0, err
	}

	return val, nil
}

func GetMetrics() (*map[string]float64, error) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := db.Query(ctx,
		"SELECT name, value FROM metrics")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metrics := make(map[string]float64)

	for rows.Next() {
		var m StorageMetric
		err := rows.Scan(&m.Name, &m.Value)
		if err != nil {
			return nil, err
		}

		metrics[m.Name] = m.Value
	}

	if rows.Err() != nil {
		return nil, err
	}

	return &metrics, nil
}

func WriteMetrics(metrics *map[string]float64) error {
	db := database.GetDB()
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

	_, err := db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}
