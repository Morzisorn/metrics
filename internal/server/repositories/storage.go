package repositories

import (
	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/server/logger"
	"github.com/morzisorn/metrics/internal/server/repositories/database"
	"github.com/morzisorn/metrics/internal/server/repositories/file"
	"github.com/morzisorn/metrics/internal/server/repositories/memory"
	"go.uber.org/zap"
)

func NewStorage(cfg config.Config) Storage {
	switch cfg.StorageType {
	case "db":
		return database.NewStorage()
	case "file":
		file, err := file.NewStorage(cfg.FileStoragePath)
		if err != nil {
			logger.Log.Panic("Incorrect file storage ", zap.Error(err))
		}
		return file
	case "memory":
		return memory.GetStorage()
	default:
		logger.Log.Panic("Incorrect storage config", zap.String("Incorrect storage type: ", cfg.StorageType))
	}
	return nil
}

type Storage interface {
	GetMetric(name string) (float64, bool)
	GetMetrics() (*map[string]float64, error)
	UpdateCounter(name string, value float64) (float64, error)
	UpdateGauge(name string, value float64) error
	WriteMetrics(*map[string]float64) error
	UpdateCounters(*map[string]float64) error
	UpdateGauges(*map[string]float64) error
}
