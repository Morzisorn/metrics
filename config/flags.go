package config

import (
	"os"

	"github.com/spf13/pflag"
)

func (c *Config) parseAgentFlags() {
	pflag.StringVarP(&c.Addr, "addr", "a", "localhost:8080", "address and port to run agent")
	pflag.StringVarP(&c.Key, "key", "k", "", "secret key")
	pflag.Float64VarP(&c.PollInterval, "poll", "p", 2, "poll interval")
	pflag.Float64VarP(&c.ReportInterval, "report", "r", 10, "report interval")

	if err := pflag.CommandLine.Parse(os.Args[1:]); err != nil {
		panic(err)
	}
}

func (c *Config) parseServerFlags() error {
	pflag.StringVarP(&c.Addr, "addr", "a", "localhost:8080", "address and port to run agent")
	pflag.StringVarP(&c.Key, "key", "k", "", "secret key")
	pflag.Int64VarP(&c.StoreInterval, "store", "i", 300, "store interval")
	pflag.StringVarP(&c.FileStoragePath, "file", "f", "storage.json", "file storage path")
	pflag.BoolVarP(&c.Restore, "restore", "r", true, "restore storage from file")
	pflag.StringVarP(&c.DBConnStr, "dbstr", "d", "", "db connection string")

	return pflag.CommandLine.Parse(os.Args[1:])
}
