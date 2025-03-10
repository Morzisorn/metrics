package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func PingDB(db *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	return db.Ping(ctx)
}

func (db *DBStorage) UpdateGauge(name string, value float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	db.mu.Lock()
	defer db.mu.Unlock()

	var val float64
	err := db.Pool.QueryRow(ctx,
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

	db.mu.Lock()
	defer db.mu.Unlock()

	var val float64

	err := db.Pool.QueryRow(ctx,
		`INSERT INTO metrics(name, value) 
		VALUES($1, $2) 
		ON CONFLICT (name) DO UPDATE 
		SET value = metrics.value + EXCLUDED.value 
		RETURNING value`,
		name, value).
		Scan(&val)

	if err != nil {
		return 0, err
	}

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

	_, err := db.Pool.Exec(ctx, query, args...)
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

	db.mu.Lock()
	defer db.mu.Unlock()

	var val float64

	err := db.Pool.QueryRow(ctx,
		"SELECT value FROM metrics WHERE name = $1", name).Scan(&val)
	if err != nil {
		return 0, false
	}

	return val, true
}

func (db *DBStorage) GetMetrics() (*map[string]float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	db.mu.Lock()
	defer db.mu.Unlock()

	rows, err := db.Pool.Query(ctx,
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

	_, err := db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (db *DBStorage) Close() {
	db.Pool.Close()
}

//func (db *DBStorage)
