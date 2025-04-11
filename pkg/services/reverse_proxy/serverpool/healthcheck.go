package serverpool

import (
	"context"
	"time"

	"github.com/letronghoangminh/reproxy/pkg/interfaces"
	"github.com/letronghoangminh/reproxy/pkg/utils"
)

func LaunchHealthCheck(ctx context.Context, sp interfaces.ServerPool) {
	t := time.NewTicker(time.Second * 20)
	utils.Logger.Info("Starting health check...")
	for {
		select {
		case <-t.C:
			go HealthCheck(ctx, sp)
		case <-ctx.Done():
			utils.Logger.Info("Closing Health Check")
			return
		}
	}
}
