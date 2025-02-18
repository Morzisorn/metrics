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

	Conf.parseFlags()

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
		Conf.parseFlags()
	}, "Expected panic when parsing unknown flag")
}

func TestGetEncFilePath(t *testing.T) {
	wd, _ := os.Getwd()
	expectedPath := filepath.Join(wd, ".env")

	actualPath := getEncFilePath()

	assert.Equal(t, expectedPath, actualPath, "Путь к .env файлу должен совпадать")
}
func TestNew(t *testing.T) {
	// Устанавливаем переменные окружения для тестов
	os.Setenv("ADDRESS", "localhost:9090")
	os.Setenv("POLL_INTERVAL", "5")
	os.Setenv("REPORT_INTERVAL", "15")
	defer func() {
		os.Unsetenv("ADDRESS")
		os.Unsetenv("POLL_INTERVAL")
		os.Unsetenv("REPORT_INTERVAL")
	}()

	// Вызываем New
	service, err := New("agent")

	// Проверяем, что нет ошибки
	assert.NoError(t, err)

	// Проверяем, что конфигурация корректно загружена
	assert.Equal(t, "localhost:9090", service.Config.Addr)
	assert.Equal(t, 5.0, service.Config.PollInterval)
	assert.Equal(t, 15.0, service.Config.ReportInterval)
}
