package backend

import (
	"context"
	"net"
	"net/url"

	"github.com/letronghoangminh/reproxy/pkg/utils"
)

func IsBackendAlive(ctx context.Context, aliveChannel chan bool, u *url.URL) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", u.Host)
	if err != nil {
		utils.Logger.Debug("Site unreachable")
		aliveChannel <- false
		return
	}
	_ = conn.Close()
	aliveChannel <- true
}
