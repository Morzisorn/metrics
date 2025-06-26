package agent

import (
	"context"

	agent "github.com/morzisorn/metrics/internal/agent/services"
	pb "github.com/morzisorn/metrics/internal/proto"
)

func (g *GRPCClient) SendMetric(m *agent.Metric) error {
	req := pb.UpdateMetricRequest{
		Metric: convertMetricToPB(m),
	}

	_, err := g.Client.UpdateMetric(context.Background(), &req)

	return err
}

func (g *GRPCClient) SendMetricsBatch(m *agent.Metrics) error {
	return nil
}

func convertMetricToPB(m *agent.Metric) *pb.Metric {
	p := pb.Metric{
		Id:    m.ID,
		Mtype: stringToPBMType(m.MType),
	}
	if m.Delta != nil {
		p.Delta = *m.Delta
	}
	if m.Value != nil {
		p.Value = *m.Value
	}
	return &p
}

func stringToPBMType(s string) pb.Metric_MType {
	switch s {
	case "GAUGE":
		return pb.Metric_GAUGE
	case "COUNTER":
		return pb.Metric_COUNTER
	default:
		return pb.Metric_GAUGE
	}
}

func (g *GRPCClient) Close() {}
