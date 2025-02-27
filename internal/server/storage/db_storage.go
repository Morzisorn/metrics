package storage

import (
	"context"
	"time"

	"github.com/morzisorn/metrics/internal/database"
)

func PingDB() error {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	return db.Ping(ctx)
}
