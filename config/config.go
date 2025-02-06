package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
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
		fmt.Println("Load .env error: ", err)
	}

	c := &Config{}

	if err := c.parseEnv(app); err != nil {
		return &Service{
			Config: *c,
		}, fmt.Errorf("Error parsing env: %v", err)
	}

	return &Service{
		Config: *c,
	}, nil
}

func getEncFilePath() string {
	_, currentFile, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(filepath.Dir(currentFile))
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
		poll, err := strconv.ParseFloat(os.Getenv("POLL_INTERVAL"), 64)
		if err != nil {
			fmt.Println("Error parsing POLL_INTERVAL: ", err)
		}
		if poll != 0 {
			c.PollInterval = poll
		}

		report, err := strconv.ParseFloat(os.Getenv("REPORT_INTERVAL"), 64)
		if err != nil {
			fmt.Println("Error parsing REPORT_INTERVAL: ", err)
		}
		if report != 0 {
			c.ReportInterval = report
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
