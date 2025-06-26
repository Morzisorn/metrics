package agent

import (
	"sync"

	"github.com/morzisorn/metrics/config"
	agent "github.com/morzisorn/metrics/internal/agent/services"
	"github.com/morzisorn/metrics/internal/server/logger"
	"go.uber.org/zap"
)

// MetricsClient is used to send single metric or batch to server
type MetricsClient interface {
	SendMetric(m *agent.Metric) error
	SendMetricsBatch(m *agent.Metrics) error
	Close()
}

func NewMetricClient(cnfg *config.Service) MetricsClient {
	switch cnfg.Config.Protocol {
	case "http":
		return NewHTTPClient(cnfg)
	case "grpc":
		return NewGRPCClient(cnfg)
	default:
		logger.Log.Panic("Create new metric client error: incorrect protocol")
		return nil
	}
}

func metricSenderJob(client MetricsClient, chIn chan agent.Metric, wg *sync.WaitGroup) {
	defer wg.Done()
	for metric := range chIn {
		err := client.SendMetric(&metric)
		if err != nil {
			logger.Log.Error("Send metric error. ",
				zap.String("Metric: ", metric.ID),
				zap.Error(err),
			)
			return
		}
	}
}

// SendMetricsByOne reads metrics from channel and manage sending them by one
func SendMetricsByOne(client MetricsClient, m *agent.Metrics) error {
	chIn := make(chan agent.Metric, len(m.Metrics))

	var wg sync.WaitGroup

	rateLimit := config.GetService().Config.RateLimit
	runWorkers(client,chIn, &wg, int(rateLimit))

	m.LoadMetricsToChan(chIn)
	close(chIn)

	wg.Wait()

	m.ResetCounter()

	return nil
}

func runWorkers(client MetricsClient, chIn chan agent.Metric, wg *sync.WaitGroup, rateLimit int) {
	for w := 0; w < rateLimit; w++ {
		wg.Add(1)
		go metricSenderJob(client,chIn, wg)
	}
}
