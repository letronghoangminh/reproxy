package matcher

import (
	"net"
	"net/http"
	"slices"
	"strings"

	"github.com/letronghoangminh/reproxy/pkg/config"
)

func MatchHandler(r *http.Request, handlers []*config.HandlerConfig) *config.HandlerConfig {
	for _, handler := range handlers {
		if len(handler.Matchers.Method) > 0 && !(slices.Contains(handler.Matchers.Method, r.Method) ||
			slices.Contains(handler.Matchers.Method, "*")) {
			continue
		}
		if handler.Matchers.Path != "" && !strings.HasPrefix(r.URL.Path, handler.Matchers.Path) {
			continue
		}
		if handler.Matchers.Headers != nil {
			for key, value := range handler.Matchers.Headers {
				if r.Header.Get(key) != value {
					continue
				}
			}
		}
		if handler.Matchers.Query != nil {
			for key, value := range handler.Matchers.Query {
				if r.URL.Query().Get(key) != value {
					continue
				}
			}
		}
		if len(handler.Matchers.ClientCIDRs) > 0 {
			ipInRange := false
			clientIp, _, _ := net.SplitHostPort(r.RemoteAddr)
			for _, cidr := range handler.Matchers.ClientCIDRs {
				_, ipNet, _ := net.ParseCIDR(cidr)
				if ipNet.Contains(net.ParseIP(clientIp)) {
					ipInRange = true
					break
				}
			}
			if !ipInRange {
				continue
			}
		}
		return handler
	}
	return nil
}
