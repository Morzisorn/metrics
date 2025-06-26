package grpc_ctrl

import (
	"context"
	"strings"

	"github.com/morzisorn/metrics/internal/models"
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

	if in.Metric.Delta == 0 && in.Metric.Value == 0 {
		return nil, status.Error(codes.InvalidArgument, "both value and delta is empty")
	}

	metric := metrics.Metric{
		Metric: models.Metric{
			ID:    in.Metric.Id,
			Delta: &in.Metric.Delta,
			Value: &in.Metric.Value,
			MType: strings.ToLower(in.Metric.GetMtype().String()),
		},
	}

	err := mc.service.UpdateMetric(&metric)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update metric error: %s", err)
	}

	logger.Log.Info("One metric successfully updated")

	return &pb.UpdateMetricResponse{
		Metric: convertMetricToPB(&metric),
	}, nil
}

func convertMetricToPB(m *metrics.Metric) *pb.Metric {
	return &pb.Metric{
		Id:    m.ID,
		Mtype: stringToPBMType(m.MType),
		Value: *m.Value,
		Delta: *m.Delta,
	}
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
