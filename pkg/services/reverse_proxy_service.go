package services

import (
	"math/rand"
	"net/http"
	"strings"

	"github.com/letronghoangminh/reproxy/pkg/config"
	"github.com/letronghoangminh/reproxy/pkg/utils"
	"go.uber.org/zap"
)

var (
	/*
		{
			"&handler1": {
				"upstream1": 1
				"upstream2": 2
			}
			"&handler2": {
				"upstream3": 1
			}
		}
	*/
	// Storing request counts for each upstream to implement load balancing
	// across multiple handlers
	// TODO: implement Mutex lock
	loadBalancingStates = map[*config.HandlerConfig]map[string]int{}
	logger *zap.Logger = zap.L()
)

func HandleReverseProxyRequest(w *http.ResponseWriter, r *http.Request, handler *config.HandlerConfig) {
	if handler.ReverseProxy.Upstreams == nil || len(handler.ReverseProxy.Upstreams) == 0 {
		http.Error(*w, "No upstreams configured", http.StatusInternalServerError)
		return
	}

	if _, ok := loadBalancingStates[handler]; !ok {
		loadBalancingStates[handler] = make(map[string]int)
		for _, upstream := range handler.ReverseProxy.Upstreams {
			loadBalancingStates[handler][upstream] = 0
		}
	}

	upstream := getNextUpstream(handler, r)

	logger.Debug("selected upstream", zap.String("upstream", upstream))

	// Create a reverse proxy to the selected upstream
	// proxy := services.NewReverseProxy(upstream)
	// proxy.ServeHTTP(*w, r)
}

func getNextUpstream(handler *config.HandlerConfig, r *http.Request) string {
	upstreams := handler.ReverseProxy.Upstreams
	var nextUpstream string 

	switch handler.ReverseProxy.LoadBalancing.Strategy {
	case "round_robin":
		for upstream, count := range loadBalancingStates[handler] {
			if count < loadBalancingStates[handler][nextUpstream] {
				nextUpstream = upstream
			}
		}
	case "weighted_round_robin":
		// TODO: implement weighted round robin
	case "random":
		nextUpstream = upstreams[rand.Intn(len(upstreams))]
	case "ip_hash":
		clientIp := strings.Split(r.RemoteAddr, ":")[0]
		ipHash := utils.Hash(clientIp)
		nextUpstream = upstreams[ipHash % uint32(len(upstreams))]
	case "uri_hash":
		uriHash := utils.Hash(r.URL.Path)
		nextUpstream = upstreams[uriHash % uint32(len(upstreams))]
	default:
		nextUpstream = upstreams[0]
	}

	loadBalancingStates[handler][nextUpstream]++

	return nextUpstream
}
