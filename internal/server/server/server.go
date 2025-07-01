package server

import (
	"context"

	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/server/repositories"
	"github.com/morzisorn/metrics/internal/server/services/health"
	"github.com/morzisorn/metrics/internal/server/services/metrics"
	"github.com/morzisorn/metrics/internal/server/services/pages"
)

type Server interface {
	Create(ms *metrics.MetricService, ps *pages.PagesService, hs *health.HealthService, storage *repositories.Storage)
	Run()
	Shutdown(ctx context.Context, idleConnsClosed chan struct{})
}

func CreateAndRun(ms *metrics.MetricService, ps *pages.PagesService, hs *health.HealthService, storage *repositories.Storage) {
	var h *HTTPServer
	var g *GRPCServer

	switch config.GetService().Config.Protocol {
	case "http":
		h = createHTTPServer(ms, ps, hs, storage)
	case "grpc":
		h = createHTTPServer(nil, ps, hs, storage)
		g = createGRPCServer(ms, storage)
	default:
		h = createHTTPServer(ms, ps, hs, storage)
	}

	go h.Run() 
	if g != nil {
		g.Run()
	}
}
