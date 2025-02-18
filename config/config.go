package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/morzisorn/metrics/internal/server/logger"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

type Config struct {
	Addr           string
	PollInterval   float64
	ReportInterval float64
}

type Service struct {
	Config Config
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

func loadEnvFile(envPath string) error {
	return godotenv.Load(envPath)
}

func (c *Config) parseEnv(app string) error {
	c.parseFlags()

	addr := os.Getenv("ADDRESS")
	if addr != "" {
		c.Addr = addr
	}
	if app == "agent" {
		p := os.Getenv("POLL_INTERVAL")
		if p != "" {
			poll, err := strconv.ParseFloat(p, 64)
			if err != nil {
				fmt.Println("Error parsing POLL_INTERVAL: ", err)
			}
			if poll != 0 {
				c.PollInterval = poll
			}
		}
		r := os.Getenv("REPORT_INTERVAL")
		if r != "" {
			report, err := strconv.ParseFloat(r, 64)
			if err != nil {
				fmt.Println("Error parsing REPORT_INTERVAL: ", err)
			}
			if report != 0 {
				c.ReportInterval = report
			}
		}
	}
	return nil
}

func (c *Config) parseFlags() {
	pflag.StringVarP(&c.Addr, "addr", "a", "localhost:8080", "address and port to run agent")
	pflag.Float64VarP(&c.PollInterval, "poll", "p", 2, "address and port to run agent")
	pflag.Float64VarP(&c.ReportInterval, "report", "r", 10, "address and port to run agent")

	if err := pflag.CommandLine.Parse(os.Args[1:]); err != nil {
		panic(err)
	}
}
