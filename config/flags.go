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
	pflag.Int64VarP(&c.RateLimit, "rate_limit", "l", 3, "rate limit")
	pflag.StringVarP(&c.CryptoKeyPath, "crypto-key", "", "", "crypto key")
	pflag.StringVarP(&c.ConfigFile, "config", "c", "config.json", "config file name")

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
	pflag.StringVarP(&c.CryptoKeyPath, "crypto-key", "", "", "crypto key")
	pflag.StringVarP(&c.ConfigFile, "config", "c", "config.json", "config file name")
	pflag.StringVarP(&c.TrustedSubnet, "trusted-subnet", "t", "127.0.0.0/8", "trusted subnet")

	return pflag.CommandLine.Parse(os.Args[1:])
}
