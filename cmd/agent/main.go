package main

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/morzisorn/metrics/config"
	client "github.com/morzisorn/metrics/internal/agent/client"
	agent "github.com/morzisorn/metrics/internal/agent/services"
)

var Service *config.Service

func RunAgent() error {
	now := time.Now()
	lastReport := time.Now()
	m := agent.Metrics{}
	c := client.HTTPClient{
		BaseURL: Service.Config.Addr,
		Client:  resty.New().SetBaseURL(Service.Config.Addr),
	}
	fmt.Println("Running agent on", Service.Config.Addr)
	for {
		if time.Since(now).Seconds() >= Service.Config.PollInterval {
			now = time.Now()
			err := m.PollMetrics()
			if err != nil {
				return err
			}
			if time.Since(lastReport).Seconds() >= Service.Config.ReportInterval {
				lastReport = time.Now()
				err := c.SendMetrics(&m)
				if err != nil {
					return err
				}
			}
		}
	}
}

func main() {
	var err error
	Service, err = config.New("agent")
	if err != nil {
		panic(err)
	}
	err = RunAgent()
	if err != nil {
		fmt.Println(err)
	}
}
