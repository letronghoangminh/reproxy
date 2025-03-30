package services

import (
	"io"
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
	loadBalancingStates             = map[*config.HandlerConfig]map[string]int{}
	logger              *zap.Logger = zap.L()
)

func HandleReverseProxyRequest(w http.ResponseWriter, r *http.Request, handler *config.HandlerConfig) {
	if handler.ReverseProxy.Upstreams == nil || len(handler.ReverseProxy.Upstreams) == 0 {
		http.Error(w, "No upstreams configured", http.StatusInternalServerError)
		return
	}

	if _, ok := loadBalancingStates[handler]; !ok {
		loadBalancingStates[handler] = make(map[string]int)
		for _, upstream := range handler.ReverseProxy.Upstreams {
			loadBalancingStates[handler][upstream] = 0
		}
	}

	addHeaders(r, handler.ReverseProxy.AddHeaders)
	removeHeaders(r, handler.ReverseProxy.RemoveHeaders)
	upstream := getNextUpstream(handler, r)

	logger.Debug("selected upstream", zap.String("upstream", upstream))

	doProxyRequest(r, w, upstream)

	loadBalancingStates[handler][upstream]++
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
	case "random":
		nextUpstream = upstreams[rand.Intn(len(upstreams))]
	case "ip_hash":
		clientIp := strings.Split(r.RemoteAddr, ":")[0]
		ipHash := utils.Hash(clientIp)
		nextUpstream = upstreams[ipHash%uint32(len(upstreams))]
	case "uri_hash":
		uriHash := utils.Hash(r.URL.Path)
		nextUpstream = upstreams[uriHash%uint32(len(upstreams))]
	default:
		nextUpstream = upstreams[0]
	}

	loadBalancingStates[handler][nextUpstream]++

	return nextUpstream
}

func removeHeaders(r *http.Request, headers []string) {
	for _, header := range headers {
		r.Header.Del(header)
	}
}

func addHeaders(r *http.Request, headers map[string]string) {
	r.Header.Add("X-Forwarded-For", r.RemoteAddr)
	r.Header.Add("X-Forwarded-Host", r.Host)
	r.Header.Add("X-Forwarded-Proto", r.URL.Scheme)

	for key, value := range headers {
		r.Header.Add(key, replaceHeaderValue(r, value))
	}
}

func replaceHeaderValue(r *http.Request, value string) string {
	replacements := map[string]string{
		"{remote_ip}":  strings.Split(r.RemoteAddr, ":")[0],
		"{scheme}":     r.URL.Scheme,
		"{host}":       r.Host,
		"{path}":       r.URL.Path,
		"{query}":      r.URL.RawQuery,
		"{method}":     r.Method,
		"{user_agent}": r.UserAgent(),
	}

	result := value
	for placeholder, replacement := range replacements {
		result = strings.Replace(result, placeholder, replacement, -1)
	}

	return result
}

func doProxyRequest(r *http.Request, w http.ResponseWriter, upstream string) {
	proxyReq, err := http.NewRequest(r.Method, upstream+r.URL.Path, r.Body)
	if err != nil {
		logger.Error("error occurred while creating proxy request", zap.Error(err))
		http.Error(w, "error occurred while creating proxy request", http.StatusInternalServerError)
		return
	}
	proxyReq.Header = r.Header

	logger.Debug("proxying request", zap.String("method", proxyReq.Method), zap.String("url", proxyReq.URL.String()))

	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		logger.Error("error occurred while sending request to upstream", zap.Error(err))
		http.Error(w, "error occurred while sending request to upstream", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	logger.Debug("received response from upstream", zap.Int("status_code", resp.StatusCode))

	if _, err := io.Copy(w, resp.Body); err != nil {
		logger.Error("error occurred while copying response body", zap.Error(err))
		http.Error(w, "error occurred while copying response body", http.StatusInternalServerError)
		return
	}
}
