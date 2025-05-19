package repositories

import (
	"testing"

	"github.com/morzisorn/metrics/config"
	"github.com/stretchr/testify/require"
)

func TestNewStorageDB(t *testing.T) {
	cfg := config.GetService("server")

	cfg.Config.StorageType = "db"

	storage := NewStorage(cfg.Config)
	require.NotNil(t, storage)
}

func TestNewStorageFile(t *testing.T) {
	cfg := config.GetService("server")

	cfg.Config.StorageType = "file"

	storage := NewStorage(cfg.Config)
	require.NotNil(t, storage)
}

func TestNewStorageMem(t *testing.T) {
	cfg := config.GetService("server")

	cfg.Config.StorageType = "memory"

	storage := NewStorage(cfg.Config)
	require.NotNil(t, storage)
}

func TestNewStorageUnknown(t *testing.T) {
	cfg := config.GetService("server")

	cfg.Config.StorageType = "unknown"

	require.Panics(t, func() { NewStorage(cfg.Config) })
}
