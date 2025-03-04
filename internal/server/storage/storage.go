package storage

import (
	"sync"

	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/server/storage/database"
	"github.com/morzisorn/metrics/internal/server/storage/file"
	"github.com/morzisorn/metrics/internal/server/storage/models"
)

var (
	instance models.Storage
	once     sync.Once
)

func GetStorage() models.Storage {
	once.Do(func() {
		service := config.GetService("server")

		if service.Config.DBConnStr != "" {
			instance = database.GetStorage()
		} else {
			instance = file.GetStorage()
		}
	})
	return instance
}
