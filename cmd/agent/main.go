package main

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/morzisorn/metrics/internal/agent"
	"github.com/morzisorn/metrics/internal/config"
)

var Conf config.Config

func RunAgent() error {
	now := time.Now()
	lastReport := time.Now()
	m := agent.Metrics{}
	client := resty.New().SetBaseURL(Conf.Addr)
	for {
		if time.Since(now).Seconds() >= Conf.PollInterval {
			now = time.Now()
			err := m.PollMetrics()
			if err != nil {
				return err
			}
			if time.Since(lastReport).Seconds() >= Conf.ReportInterval {
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
	Conf.Init()
	err := RunAgent()
	if err != nil {
		fmt.Println(err)
	}
}
