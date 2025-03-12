package database

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/server/logger"
	"go.uber.org/zap"
)

var (
	//instanceStorage models.Storage
	//onceStorage     sync.Once

	instancePool *pgxpool.Pool
	oncePool     sync.Once
)

type DBStorage struct {
	Pool *pgxpool.Pool
	mu   sync.RWMutex
}

func NewStorage() *DBStorage {
	return &DBStorage{Pool: GetDBPool()}
}

func GetDBPool() *pgxpool.Pool {
	oncePool.Do(func() {
		var err error
		s := config.GetService("server")
		instancePool, err = pgxpool.New(context.Background(), s.Config.DBConnStr)
		if err != nil {
			logger.Log.Panic("Unable to connect to database: ", zap.Error(err))
		}

		err = createTables(instancePool)
		if err != nil {
			logger.Log.Panic("Unable to create database tables: ", zap.Error(err))
		}
	})

	return instancePool
}

func createTables(db *pgxpool.Pool) error {
	rootDir, err := config.GetProjectRoot()
	if err != nil {
		return err
	}
	filepath := filepath.Join(rootDir, "internal", "server", "repositories", "database", "db_structure.sql")

	script, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	_, err = db.Exec(context.Background(), string(script))
	if err != nil {
		return err
	}

	return nil
}
