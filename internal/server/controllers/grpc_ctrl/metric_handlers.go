package grpc_ctrl

import (
	"context"

	pb "github.com/morzisorn/metrics/internal/proto"
	"github.com/morzisorn/metrics/internal/server/logger"
	"github.com/morzisorn/metrics/internal/server/services/metrics"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MetricController struct {
	pb.UnimplementedMetricControllerServer
	service *metrics.MetricService
}

// NewMetricController receives Metric service and creates Metric controller, returns pointer.
func NewMetricController(service *metrics.MetricService) *MetricController {
	return &MetricController{
		service: service,
	}
}

func (mc *MetricController) UpdateMetric(ctx context.Context, in *pb.UpdateMetricRequest) (*pb.UpdateMetricResponse, error) {
	if in.Metric.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "metric id is empty")
	}

	metric := metrics.Metric{Metric: *pb.ConvertMetricFromPB(in.Metric)}

	err := mc.service.UpdateMetric(&metric)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update metric error: %s", err)
	}

	logger.Log.Info("One metric successfully updated")

	return &pb.UpdateMetricResponse{
		Metric: pb.ConvertMetricToPB(&metric.Metric),
	}, nil
}

func (mc *MetricController) UpdateMetrics(ctx context.Context, in *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	ms := make([]metrics.Metric, len(in.Metrics))
	for i, mPB := range in.Metrics {
		m := metrics.Metric{Metric: *pb.ConvertMetricFromPB(mPB)}
		ms[i] = m
	}

	err := mc.service.UpdateMetrics(&ms)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update metrics error: %s", err)
	}

	logger.Log.Info("Batch of metrics successfully updated")

	return &pb.UpdateMetricsResponse{}, nil
}
