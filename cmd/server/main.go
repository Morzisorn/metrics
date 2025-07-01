package main

import (
	"go.uber.org/zap"

	_ "net/http/pprof"

	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/server/logger"
	"github.com/morzisorn/metrics/internal/server/repositories"
	"github.com/morzisorn/metrics/internal/server/server"
	"github.com/morzisorn/metrics/internal/server/services/health"
	"github.com/morzisorn/metrics/internal/server/services/metrics"
	"github.com/morzisorn/metrics/internal/server/services/pages"
)

// buildVersion represents the version of the application.
// It is injected during build using -ldflags.
// Example: go build -ldflags "-X main.buildVersion=1.0.0"
var buildVersion string

// buildDate represents the build date of the application.
// It is injected during build using -ldflags.
// Example: go build -ldflags "-X main.buildDate=2025-05-29"
var buildDate string

// buildCommit represents the commit from which the application was built.
// It is injected during build using -ldflags.
// Example: go build -ldflags "-X main.buildCommit=abc123"
var buildCommit string

var (
	cnfg *config.Service
)

func main() {
	if err := logger.Init(); err != nil {
		logger.Log.Fatal("Init logger error: ", zap.Error(err))
	}

	config.PrintMetaInfo(buildVersion, buildDate, buildCommit)
	cnfg = config.GetService("server")

	storage := repositories.NewStorage(cnfg.Config)
	metricsService := metrics.NewMetricService(storage)
	pagesService := pages.NewPagesService(metricsService)
	healthService := health.NewHealthService(storage)

	if cnfg.Config.StorageType == "file" && cnfg.Config.Restore {
		if err := metricsService.LoadMetricsFromFile(); err != nil {
			logger.Log.Panic("Error loading metrics", zap.Error(err))
		}
	}

	server.CreateAndRun(metricsService, pagesService, healthService, &storage)
}
