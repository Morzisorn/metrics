package main

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestParseAgentFlagsOK(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{"cmd", "--a", "localhost:9000", "--p", "5", "--r", "15"}

	AgentConf = AgentConfig{}

	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)

	parseAgentFlags()

	assert.Equal(t, "localhost:9000", AgentConf.Addr)
	assert.Equal(t, 5.0, AgentConf.PollInterval)
	assert.Equal(t, 15.0, AgentConf.ReportInterval)
}

func TestParseAgentFlagsUnknown(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{"cmd", "--z", "localhost:9000"}

	AgentConf = AgentConfig{}

	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)

	assert.Panics(t, func() {
		parseAgentFlags()
	}, "Expected panic when parsing unknown flag")
}
