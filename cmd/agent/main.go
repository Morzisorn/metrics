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
	var err error
	Service = config.GetService("agent")
	err = RunAgent()
	if err != nil {
		fmt.Println(err)
	}
}
