package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/morzisorn/metrics/internal/server/logger"
	"go.uber.org/zap"
)

func loadEnvFile(envPath string) error {
	return godotenv.Load(envPath)
}

func (c *Config) parseEnv(app string) error {
	switch app {
	case "agent":
		c.parseAgentEnvs()
	case "server":
		c.parseServerEnvs()
	}

	return nil
}

func (c *Config) parseAgentEnvs() {
	c.parseAgentFlags()

	addr := os.Getenv("ADDRESS")
	if addr != "" {
		c.Addr = addr
	}

	f, err := getEnvFloat("POLL_INTERVAL")
	if err == nil && f != 0 {
		c.PollInterval = f
	}

	f, err = getEnvFloat("REPORT_INTERVAL")
	if err == nil && f != 0 {
		c.ReportInterval = f
	}
}

func (c *Config) parseServerEnvs() {
	err := c.parseServerFlags()
	if err != nil {
		logger.Log.Panic("Parse flags error ", zap.Error(err))
	}

	addr := os.Getenv("ADDRESS")
	if addr != "" {
		c.Addr = addr
	}

	i, err := getEnvInt("STORE_INTERVAL")
	if err == nil {
		c.StoreInterval = i
	}

	s, err := getEnvString("FILE_STORAGE_PATH")
	if err == nil && s != "" {
		c.FileStoragePath = s
	}

	b, err := getEnvBool("RESTORE")
	if err == nil {
		c.Restore = b
	}

	d, err := getEnvString("DATABASE_DSN")
	if err == nil {
		c.DBConnStr = d
	}
}

func getEnvFloat(key string) (float64, error) {
	env := os.Getenv(key)
	if env != "" {
		return strconv.ParseFloat(env, 64)
	}
	return 0, fmt.Errorf("env %s not found", key)
}

func getEnvInt(key string) (int64, error) {
	env := os.Getenv(key)
	if env != "" {
		return strconv.ParseInt(env, 10, 64)
	}
	return 0, fmt.Errorf("env %s not found", key)
}

func getEnvString(key string) (string, error) {
	env := os.Getenv(key)
	if env != "" {
		return env, nil
	}
	return "", fmt.Errorf("env %s not found", key)
}

func getEnvBool(key string) (bool, error) {
	env := os.Getenv(key)
	if env != "" {
		return strconv.ParseBool(env)
	}
	return false, fmt.Errorf("env %s not found", key)
}
