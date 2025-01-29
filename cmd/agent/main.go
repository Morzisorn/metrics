package main

import (
	"fmt"
	"time"

	"github.com/morzisorn/metrics/internal/agent"
)

const (
	pollInterval   = 2
	reportInterval = 10
)

func RunAgent() error {
	now := time.Now()
	lastReport := time.Now()
	m := agent.Metrics{}
	for {
		if time.Since(now).Seconds() >= pollInterval {
			now = time.Now()
			err := m.PollMetrics()
			if err != nil {
				return err
			}
			if time.Since(lastReport) >= reportInterval {
				lastReport = time.Now()
				err := m.SendMetrics()
				if err != nil {
					return err
				}
			}
		}
	}
}

func main() {
	err := RunAgent()
	if err != nil {
		fmt.Println(err)
	}
}
