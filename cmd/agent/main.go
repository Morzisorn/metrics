package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
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
	c := client.NewMetricClient(Service)
	logger.Log.Info("Running agent.", zap.String("Address: ", Service.Config.Addr))
	idleConnsClosed := make(chan struct{})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

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
					err := client.SendMetricsByOne(c, &m)
					if err != nil {
						return err
					}

					time.Sleep(1 * time.Second)

					err = c.SendMetricsBatch(&m)
					if err != nil {
						return err
					}
				}
				select {
				case <-quit:
					shutdown(idleConnsClosed, c)
					<-idleConnsClosed
					return nil
				default:
				}
			}
		}
	}
}

func shutdown(idleConnsClosed chan struct{}, c client.MetricsClient) {
	logger.Log.Info("Shutdown agent")
	c.Close()
	close(idleConnsClosed)
}

func main() {
	config.PrintMetaInfo(buildVersion, buildDate, buildCommit)
	err := logger.Init()
	if err != nil {
		fmt.Println(err)
	}
	Service = config.GetService("agent")
	err = RunAgent()
	if err != nil {
		fmt.Println(err)
	}
}
