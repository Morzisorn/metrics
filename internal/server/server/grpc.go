package server

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/morzisorn/metrics/config"
	pb "github.com/morzisorn/metrics/internal/proto"
	"github.com/morzisorn/metrics/internal/server/controllers/grpcapi"
	"github.com/morzisorn/metrics/internal/server/logger"
	"github.com/morzisorn/metrics/internal/server/repositories"
	"github.com/morzisorn/metrics/internal/server/services/metrics"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	Server  *grpc.Server
	Storage *repositories.Storage
	Listen  net.Listener
}

func createGRPCServer(ms *metrics.MetricService, storage *repositories.Storage) *GRPCServer {
	metricController := grpcapi.NewMetricController(ms)
	cnfg := config.GetService()
	listen, err := net.Listen("tcp", cnfg.Config.AddrGRPC)
	if err != nil {
		logger.Log.Fatal("create listener error", zap.Error(err))
	}

	s := grpc.NewServer()
	pb.RegisterMetricControllerServer(s, metricController)

	return &GRPCServer{s, storage, listen}
}

func (s *GRPCServer) Run() {
	logger.Log.Info("Run grpc server")

	idleConnsClosed := make(chan struct{})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		s.Shutdown(ctx, idleConnsClosed)
	}()

	if err := s.Server.Serve(s.Listen); err != nil {
		logger.Log.Fatal("Failed to run grpc server", zap.Error(err))
	}

	<-idleConnsClosed

	logger.Log.Info("Server shutted down gracefully")
}

func (s *GRPCServer) Shutdown(ctx context.Context, idleConnsClosed chan struct{}) {
	logger.Log.Info("Shutdown server")

	(*s.Storage).Close()

	s.Server.GracefulStop()

	close(idleConnsClosed)
}
