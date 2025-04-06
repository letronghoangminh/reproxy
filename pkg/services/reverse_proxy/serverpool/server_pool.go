package serverpool

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/letronghoangminh/reproxy/pkg/services/reverse_proxy/backend"
	"github.com/letronghoangminh/reproxy/pkg/utils"
	"go.uber.org/zap"
)

type ServerPool interface {
	GetBackends() []backend.Backend
	GetNextValidPeer(r *http.Request) backend.Backend
	AddBackend(backend.Backend)
	GetServerPoolSize() int
}

func HealthCheck(ctx context.Context, s ServerPool) {
	aliveChannel := make(chan bool, 1)

	for _, b := range s.GetBackends() {
		requestCtx, stop := context.WithTimeout(ctx, 10 * time.Second)
		defer stop()
		status := "up"
		go backend.IsBackendAlive(requestCtx, aliveChannel, b.GetURL())

		select {
		case <-ctx.Done():
			utils.Logger.Info("Gracefully shutting down health check")
			return
		case alive := <-aliveChannel:
			b.SetAlive(alive)
			if !alive {
				status = "down"
			}
		}
		utils.Logger.Debug(
			"URL Status",
			zap.String("URL", b.GetURL().String()),
			zap.String("status", status),
		)
	}
}

func NewServerPool(strategy LBStrategy) (ServerPool, error) {
	switch strategy {
	case RoundRobin:
		return &roundRobinServerPool{
			backends: make([]backend.Backend, 0),
			current:  0,
		}, nil
	case LeastConnections:
		return &lcServerPool{
			backends: make([]backend.Backend, 0),
		}, nil
	case Random:
		return &randomServerPool{
			backends: make([]backend.Backend, 0),
		}, nil
	case IPHash:
		return &ipServerPool{
			backends: make([]backend.Backend, 0),
		}, nil
	case URIHash:
		return &uriServerPool{
			backends: make([]backend.Backend, 0),
		}, nil
	case Sticky:
		return &stickyServerPool{
			backends: make([]backend.Backend, 0),
			current:  0,
		}, nil
	default:
		return nil, fmt.Errorf("Invalid strategy")
	}
}
