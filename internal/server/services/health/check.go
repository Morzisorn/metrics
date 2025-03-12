package health

import (
	"github.com/morzisorn/metrics/internal/server/repositories"
	"github.com/morzisorn/metrics/internal/server/repositories/database"
)

type HealthService struct {
	storage repositories.Storage
}

func NewHealthService(storage repositories.Storage) *HealthService {
	return &HealthService{storage: storage}
}

func PingDB() error {
	return database.PingDB(database.NewStorage())
}
