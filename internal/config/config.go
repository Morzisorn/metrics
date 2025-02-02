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
	pflag.StringVarP(&c.Addr, "addr", "a", "localhost:8080", "address and port to run agent")
	pflag.Float64VarP(&c.PollInterval, "poll", "p", 2, "address and port to run agent")
	pflag.Float64VarP(&c.ReportInterval, "report", "r", 10, "address and port to run agent")

	if err := pflag.CommandLine.Parse(os.Args[1:]); err != nil {
		panic(err)
	}
}

func (c *Config) Init(app string) error {
	_, currentFile, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(filepath.Dir(filepath.Dir(currentFile)))
	envPath := filepath.Join(basePath, "cmd", app, ".env")

	err := godotenv.Load(envPath)
	if err != nil {
		fmt.Println("Load .env error: ", err)
	}

	c.ParseFlags()

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
