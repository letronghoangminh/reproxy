package backend

import (
	"context"
	"net"
	"net/url"

	"go.uber.org/zap"
)

var (
	logger *zap.Logger = zap.L()
)

func IsBackendAlive(ctx context.Context, aliveChannel chan bool, u *url.URL) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", u.Host)
	if err != nil {
		logger.Debug("Site unreachable", zap.Error(err))
		aliveChannel <- false
		return
	}
	_ = conn.Close()
	aliveChannel <- true
}
