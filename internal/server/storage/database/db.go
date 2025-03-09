package database

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/server/logger"
	"github.com/morzisorn/metrics/internal/server/storage/models"
	"go.uber.org/zap"
)

var (
	instanceStorage models.Storage
	onceStorage     sync.Once

	instanceDB *pgx.Conn
	onceDB     sync.Once
)

type DBStorage struct {
	DB *pgx.Conn
	mu sync.Mutex
}

func GetStorage() models.Storage {
	onceStorage.Do(func() {
		instanceStorage = &DBStorage{
			DB: GetDB(),
		}
	})
	return instanceStorage
}

func GetDB() *pgx.Conn {
	onceDB.Do(func() {
		var err error
		s := config.GetService("server")
		instanceDB, err = pgx.Connect(context.Background(), s.Config.DBConnStr)
		if err != nil {
			logger.Log.Panic("Unable to connect to database: ", zap.Error(err))
		}

		err = createTables(instanceDB)
		if err != nil {
			logger.Log.Panic("Unable to create database tables: ", zap.Error(err))
		}
	})

	return instanceDB
}

func createTables(db *pgx.Conn) error {
	rootDir, err := config.GetProjectRoot()
	if err != nil {
		return err
	}
	filepath := filepath.Join(rootDir, "internal", "server", "storage", "database", "db_structure.sql")

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
