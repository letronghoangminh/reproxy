package serverpool

import (
	"context"
	"time"

	"go.uber.org/zap"
)

func LaunchHealthCheck(ctx context.Context, sp ServerPool) {
	logger := zap.L()

	t := time.NewTicker(time.Second * 20)
	logger.Info("Starting health check...")
	for {
		select {
		case <-t.C:
			go HealthCheck(ctx, sp)
		case <-ctx.Done():
			logger.Info("Closing Health Check")
			return
		}
	}
}
