package services

import (
	"context"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/letronghoangminh/reproxy/pkg/config"
	"github.com/letronghoangminh/reproxy/pkg/services/reverse_proxy/backend"
	loadbalancer "github.com/letronghoangminh/reproxy/pkg/services/reverse_proxy/load_balancer"
	"github.com/letronghoangminh/reproxy/pkg/services/reverse_proxy/serverpool"
	"go.uber.org/zap"
)

// Reverse proxy implementation credits to https://github.com/leonardo5621/golang-load-balancer

var (
	loadBalancers = map[*config.HandlerConfig]loadbalancer.LoadBalancer{}
	logger              *zap.Logger = zap.L()
)

func StartLoadBalancers(handlers []*config.HandlerConfig) {
	for _, handler := range handlers {
		serverPool, err := serverpool.NewServerPool(serverpool.GetLBStrategy(handler.ReverseProxy.LoadBalancing.Strategy))
		if err != nil {
			logger.Error("error occurred while creating server pool", zap.Error(err))
			return
		}
	
		loadBalancer := loadbalancer.NewLoadBalancer(serverPool)

		for _, u := range handler.ReverseProxy.Upstreams {
			endpoint, err := url.Parse(u)
			if err != nil {
				logger.Fatal(err.Error(), zap.String("URL", u))
			}

			rp := httputil.NewSingleHostReverseProxy(endpoint)

			backendServer := backend.NewBackend(endpoint, rp)
			rp.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
				logger.Error("error handling the request",
					zap.String("host", endpoint.Host),
					zap.Error(e),
				)
				backendServer.SetAlive(false)
	
				if !loadbalancer.AllowRetry(request) {
					logger.Info(
						"Max retry attempts reached, terminating",
						zap.String("address", request.RemoteAddr),
						zap.String("path", request.URL.Path),
					)
					http.Error(writer, "Service not available", http.StatusServiceUnavailable)
					return
				}
	
				logger.Info(
					"Attempting retry",
					zap.String("address", request.RemoteAddr),
					zap.String("URL", request.URL.Path),
					zap.Bool("retry", true),
				)
				loadBalancer.Serve(
					writer,
					request.WithContext(
						context.WithValue(request.Context(), loadbalancer.RETRY_ATTEMPTED, true),
					),
				)
			}
	
			serverPool.AddBackend(backendServer)
		}

		loadBalancers[handler] = loadBalancer
	}
}

func HandleReverseProxyRequest(w http.ResponseWriter, r *http.Request, handler *config.HandlerConfig) {
	if handler.ReverseProxy.Upstreams == nil || len(handler.ReverseProxy.Upstreams) == 0 {
		http.Error(w, "No upstreams configured", http.StatusInternalServerError)
		return
	}

	loadBalancer := loadBalancers[handler]
	if loadBalancer == nil {
		http.Error(w, "Load balancer not found", http.StatusInternalServerError)
		return
	}

	addHeaders(r, handler.ReverseProxy.AddHeaders)
	removeHeaders(r, handler.ReverseProxy.RemoveHeaders)

	if handler.Path != "" {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, handler.Path)
	}

	loadBalancer.Serve(w, r)
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
		"{remote_ip}": func() string {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				return r.RemoteAddr
			}
			return host
		}(),
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
