package config

import (
	"os"

	"github.com/spf13/pflag"
)

type Config struct {
	Addr           string
	PollInterval   float64
	ReportInterval float64
}

func (c *Config) ParseFlags() {
	pflag.StringVarP(&c.Addr, "addr", "a", "[::]:8080", "address and port to run agent")
	pflag.Float64VarP(&c.PollInterval, "poll", "p", 2, "address and port to run agent")
	pflag.Float64VarP(&c.ReportInterval, "report", "r", 10, "address and port to run agent")

	if err := pflag.CommandLine.Parse(os.Args[1:]); err != nil {
		panic(err)
	}
}
