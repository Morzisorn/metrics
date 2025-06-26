package agent

import (
	"context"

	agent "github.com/morzisorn/metrics/internal/agent/services"
	pb "github.com/morzisorn/metrics/internal/proto"
)

func (g *GRPCClient) SendMetric(m *agent.Metric) error {
	req := pb.UpdateMetricRequest{
		Metric: pb.ConvertMetricToPB(&m.Metric),
	}

	_, err := g.Client.UpdateMetric(context.Background(), &req)

	return err
}

func (g *GRPCClient) SendMetricsBatch(m *agent.Metrics) error {
	metricsPB := make([]*pb.Metric, len(m.Metrics))
	var i int
	m.Mu.RLock()
	for _, metric := range m.Metrics {
		metricsPB[i] = pb.ConvertMetricToPB(&metric.Metric)
		i++
	}
	m.Mu.RUnlock()

	req := pb.UpdateMetricsRequest{
		Metrics: metricsPB,
	}

	_, err := g.Client.UpdateMetrics(context.Background(), &req)
	return err
}

func (g *GRPCClient) Close() {}
