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

func (c *Config) ParseFlags() {
	if c.Addr == "" {
		pflag.StringVarP(&c.Addr, "addr", "a", "localhost:8080", "address and port to run agent")
	}
	if c.PollInterval == 0 {
		pflag.Float64VarP(&c.PollInterval, "poll", "p", 2, "address and port to run agent")
	}
	if c.ReportInterval == 0 {
		pflag.Float64VarP(&c.ReportInterval, "report", "r", 10, "address and port to run agent")
	}

	if err := pflag.CommandLine.Parse(os.Args[1:]); err != nil {
		panic(err)
	}
}

func (c *Config) Init() error {
	_, currentFile, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(filepath.Dir(filepath.Dir(currentFile)))
	envPath := filepath.Join(basePath, "internal", "config", ".env")

	err := godotenv.Load(envPath)
	if err != nil {
		fmt.Println("Load .env error: ", err)
	}

	c.Addr = os.Getenv("ADDRESS")
	c.PollInterval, err = strconv.ParseFloat(os.Getenv("POLL_INTERVAL"), 64)
	if err != nil {
		fmt.Println("Error parsing POLL_INTERVAL: ", err)
	}
	c.ReportInterval, err = strconv.ParseFloat(os.Getenv("REPORT_INTERVAL"), 64)
	if err != nil {
		fmt.Println("Error parsing REPORT_INTERVAL: ", err)
	}

	c.ParseFlags()
	return nil
}
