package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestParseFlagsOK(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{"cmd", "-a", "localhost:9000", "-p", "5", "-r", "15"}

	Conf := Config{}

	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)

	Conf.parseAgentFlags()

	assert.Equal(t, "localhost:9000", Conf.Addr)
	assert.Equal(t, 5.0, Conf.PollInterval)
	assert.Equal(t, 15.0, Conf.ReportInterval)
}

func TestParseFlagsUnknown(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{"cmd", "-z", "localhost:9000"}

	Conf := Config{}

	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)

	assert.Panics(t, func() {
		Conf.parseAgentFlags()
	}, "Expected panic when parsing unknown flag")
}

func TestGetEncFilePath(t *testing.T) {
	wd, _ := os.Getwd()
	expectedPath := filepath.Join(wd, ".env")

	actualPath := getEncFilePath()

	assert.Equal(t, expectedPath, actualPath, "Путь к .env файлу должен совпадать")
}
