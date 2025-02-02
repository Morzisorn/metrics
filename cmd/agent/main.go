package main

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/morzisorn/metrics/internal/agent"
)

type AgentConfig struct {
	Addr           string
	PollInterval   float64
	ReportInterval float64
}

var AgentConf AgentConfig

func RunAgent() error {
	now := time.Now()
	lastReport := time.Now()
	m := agent.Metrics{}
	client := resty.New().SetBaseURL(AgentConf.Addr)
	for {
		if time.Since(now).Seconds() >= AgentConf.PollInterval {
			now = time.Now()
			err := m.PollMetrics()
			if err != nil {
				return err
			}
			if time.Since(lastReport).Seconds() >= AgentConf.ReportInterval {
				lastReport = time.Now()
				err := m.SendMetrics(client)
				if err != nil {
					return err
				}
			}
		}
	}
}

func main() {
	parseAgentFlags()
	err := RunAgent()
	if err != nil {
		fmt.Println(err)
	}
}
