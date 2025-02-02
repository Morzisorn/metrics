package main

import (
	"os"

	"github.com/spf13/pflag"
)

func parseAgentFlags() {
	pflag.StringVar(&AgentConf.Addr, "a", ":8080", "address and port to run agent")
	pflag.Float64Var(&AgentConf.PollInterval, "p", 2, "address and port to run agent")
	pflag.Float64Var(&AgentConf.ReportInterval, "r", 10, "address and port to run agent")

	if err := pflag.CommandLine.Parse(os.Args[1:]); err != nil {
		panic(err)
	}
}
