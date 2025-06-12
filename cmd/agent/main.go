package main

import (
	"fmt"
	"time"

	"github.com/morzisorn/metrics/config"
	client "github.com/morzisorn/metrics/internal/agent/client"
	agent "github.com/morzisorn/metrics/internal/agent/services"
	"github.com/morzisorn/metrics/internal/server/logger"
	"go.uber.org/zap"
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

var Service *config.Service

func RunAgent() error {
	now := time.Now()
	lastReport := time.Now()
	m := agent.Metrics{}
	c := client.NewClient(Service)
	logger.Log.Info("Running agent.", zap.String("Address: ", Service.Config.Addr))
	for {
		if time.Since(now).Seconds() >= Service.Config.PollInterval {
			now = time.Now()
			err := m.PollMetrics()
			if err != nil {
				return err
			}

			if time.Since(lastReport).Seconds() >= Service.Config.ReportInterval {
				if len(m.Metrics) > 0 {
					lastReport = time.Now()
					err := c.SendMetricsByOne(&m)
					if err != nil {
						return err
					}

					time.Sleep(1 * time.Second)

					err = c.SendMetricsBatch(&m)
					if err != nil {
						return err
					}
				}
			}
		}
	}
}

func main() {
	config.PrintMetaInfo(buildVersion, buildDate, buildCommit)
	var err error
	Service = config.GetService("agent")
	err = RunAgent()
	if err != nil {
		fmt.Println(err)
	}
}
