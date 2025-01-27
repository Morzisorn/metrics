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

func main() {
	now := time.Now()
	lastReport := time.Now()
	m := agent.Metrics{}
	for {
		if time.Since(now).Seconds() >= pollInterval {
			now = time.Now()
			err := m.PollAllMetrics()
			if err != nil {
				fmt.Println(err)
			}
			if time.Since(lastReport) >= reportInterval {
				lastReport = time.Now()
				err := m.SendAllMetrics()
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}
