package database

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/server/logger"
	"go.uber.org/zap"
)

var (
	instance *pgx.Conn
	once     sync.Once
)

func GetDB() *pgx.Conn {
	once.Do(func() {
		var err error
		s := config.GetService("server")
		instance, err = pgx.Connect(context.Background(), s.Config.DBConnStr)
		if err != nil {
			logger.Log.Panic("Unable to connect to database: ", zap.Error(err))
		}
	})
	return instance
}

func CloseDB() {
	if instance != nil {
		err := instance.Close(context.Background())
		if err != nil {
			logger.Log.Panic("DB close error: ", zap.Error(err))
		}
	}
}

