package agent

import (
	"time"

	"github.com/morzisorn/metrics/config"
	pb "github.com/morzisorn/metrics/internal/proto"
	"github.com/morzisorn/metrics/internal/server/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// HTTPClient contains pointer to resty client, server base url and delays for retrier
type GRPCClient struct {
	BaseURL     string
	Client      pb.MetricControllerClient
	retryDelays []time.Duration
}

// NewGRPCClient creates new pointer to GRPCClient based on config
func NewGRPCClient(s *config.Service) *GRPCClient {
	conn, err := grpc.NewClient("127.0.0.1:8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(1024*1024*16), // 16MB
			grpc.MaxCallSendMsgSize(1024*1024*16), // 16MB
		),
	)
	if err != nil {
		logger.Log.Fatal("create grpc client error: ", zap.Error(err))
	}
	return &GRPCClient{
		BaseURL:     s.Config.Addr,
		Client:      pb.NewMetricControllerClient(conn),
		retryDelays: s.Config.RetryDelays,
	}
}
