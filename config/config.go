package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/morzisorn/metrics/internal/server/logger"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

type Config struct {
	Addr           string
	PollInterval   float64
	ReportInterval float64

	StoreInterval   int64
	FileStoragePath string
	Restore         bool
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

func getProjectRoot() (string, error) {
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
	basePath, err := getProjectRoot()
	if err != nil {
		logger.Log.Error("Error getting project root ", zap.Error(err))
		return ".env"
	}
	return filepath.Join(basePath, "config", ".env")
}

func (c *Config) parseAgentFlags() {
	pflag.StringVarP(&c.Addr, "addr", "a", "localhost:8080", "address and port to run agent")
	pflag.Float64VarP(&c.PollInterval, "poll", "p", 2, "poll interval")
	pflag.Float64VarP(&c.ReportInterval, "report", "r", 10, "report interval")

	if err := pflag.CommandLine.Parse(os.Args[1:]); err != nil {
		panic(err)
	}
}

func (c *Config) parseServerFlags() {
	pflag.StringVarP(&c.Addr, "addr", "a", "localhost:8080", "address and port to run agent")
	pflag.Int64VarP(&c.StoreInterval, "store", "i", 300, "store interval")
	pflag.StringVarP(&c.FileStoragePath, "file", "f", "storage.json", "file storage path")
	pflag.BoolVarP(&c.Restore, "restore", "r", true, "restore storage from file")

	if err := pflag.CommandLine.Parse(os.Args[1:]); err != nil {
		panic(err)
	}
}
