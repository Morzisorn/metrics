package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/morzisorn/metrics/internal/server/logger"
	"go.uber.org/zap"
)

type Config struct {
	//Common
	Addr string

	//Agent
	PollInterval   float64
	ReportInterval float64

	//Server
	StoreInterval   int64
	FileStoragePath string
	Restore         bool
	DBConnStr       string
}

type Service struct {
	Config Config
}

var (
	instance *Service
	once     sync.Once
)

func GetService(app string) *Service {
	once.Do(func() {
		var err error
		instance, err = New(app)
		if err != nil {
			logger.Log.Error("Error getting service", zap.Error(err))
		}
	})
	return instance
}

func New(app string) (*Service, error) {
	envPath := getEncFilePath()
	if err := loadEnvFile(envPath); err != nil {
		fmt.Printf("Load .env error: %v. Env path: %s\n", err, envPath)
	}

	c := &Config{}

	if err := c.parseEnv(app); err != nil {
		return &Service{
			Config: *c,
		}, fmt.Errorf("error parsing env: %v", err)
	}

	return &Service{
		Config: *c,
	}, nil
}

func GetProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		}

		parent := filepath.Dir(wd)
		if parent == wd {
			return "", fmt.Errorf("project root not found")
		}
		wd = parent
	}
}

func getEncFilePath() string {
	basePath, err := GetProjectRoot()
	if err != nil {
		logger.Log.Error("Error getting project root ", zap.Error(err))
		return ".env"
	}
	return filepath.Join(basePath, "config", ".env")
}
