// Package proxy provides functionality to handle reverse proxy requests and load balancing.
package proxy

import (
	"context"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/letronghoangminh/reproxy/pkg/config"
	"github.com/letronghoangminh/reproxy/pkg/interfaces"
	"github.com/letronghoangminh/reproxy/pkg/services/dns"
	"github.com/letronghoangminh/reproxy/pkg/services/proxy/backend"
	loadbalancer "github.com/letronghoangminh/reproxy/pkg/services/proxy/load_balancer"
	serverpool "github.com/letronghoangminh/reproxy/pkg/services/proxy/server_pool"
	"github.com/letronghoangminh/reproxy/pkg/utils"
)

// Reverse proxy implementation credits to https://github.com/leonardo5621/golang-load-balancer

var (
	loadBalancers = map[*config.HandlerConfig]interfaces.LoadBalancer{}
)

func StartLoadBalancers(ctx context.Context, handlers []*config.HandlerConfig) {
	for _, handler := range handlers {
		serverPool, err := serverpool.NewServerPool(serverpool.GetLBStrategy(handler.ReverseProxy.LoadBalancing.Strategy))
		if err != nil {
			utils.Logger.Error("error occurred while creating server pool", "error", err)
			return
		}

		loadBalancer := loadbalancer.NewLoadBalancer(serverPool)

		staticUpstreams := handler.ReverseProxy.Upstreams.Static
		dynamicUpstreams, dnsErr := dns.GetDynamicUpstreams(handler.ReverseProxy.Upstreams.Dynamic)
		if dnsErr != nil {
			utils.Logger.Error("error resolving dynamic upstreams", "error", dnsErr)
			dynamicUpstreams = []string{}
		}

		for _, u := range append(staticUpstreams, dynamicUpstreams...) {
			endpoint, err := url.Parse(u)
			if err != nil {
				utils.Logger.Fatal(err.Error(), "URL", u)
			}

			rp := httputil.NewSingleHostReverseProxy(endpoint)

			backendServer := backend.NewBackend(endpoint, rp)
			rp.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
				utils.Logger.Debug("error handling the request",
					"host", endpoint.Host,
				)
				backendServer.SetAlive(false)

				if !loadbalancer.AllowRetry(request, handler.ReverseProxy.LoadBalancing.Retries) {
					utils.Logger.Info(
						"Max retry attempts reached, terminating",
						"address", request.RemoteAddr,
						"path", request.URL.Path,
					)
					http.Error(writer, "Service not available", http.StatusServiceUnavailable)
					return
				}

				currentCount, ok := request.Context().Value(loadbalancer.Count).(int)
				if !ok {
					currentCount = 0
				}

				utils.Logger.Info(
					"Attempting retry",
					"address", request.RemoteAddr,
					"URL", request.URL.Path,
					"retry", true,
					"retry_count", currentCount+1,
				)

				sleepDuration := handler.ReverseProxy.LoadBalancing.TryInterval
				if sleepDuration == 0 {
					sleepDuration = 5
				}
				time.Sleep(time.Duration(sleepDuration) * time.Second)

				loadBalancer.Serve(
					writer,
					request.WithContext(
						context.WithValue(request.Context(), loadbalancer.Count, currentCount+1),
					),
				)
			}

			serverPool.AddBackend(backendServer)
		}

		go serverpool.LaunchHealthCheck(ctx, serverPool)

		loadBalancers[handler] = loadBalancer
	}
}

func HandleReverseProxyRequest(w http.ResponseWriter, r *http.Request, handler *config.HandlerConfig) {
	loadBalancer := loadBalancers[handler]
	if loadBalancer == nil {
		http.Error(w, "Load balancer not found", http.StatusInternalServerError)
		return
	}

	addHeaders(r, handler.ReverseProxy.AddHeaders)
	removeHeaders(r, handler.ReverseProxy.RemoveHeaders)

	if handler.Matchers.Path != "" {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, handler.Matchers.Path)
	}

	rewritePath(r, handler.ReverseProxy.Rewrite)

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
		result = strings.ReplaceAll(result, placeholder, replacement)
	}

	return result
}

func rewritePath(r *http.Request, rewrite string) {
	if rewrite == "" {
		return
	}

	rewrite = strings.TrimPrefix(rewrite, "/")
	rewrite = strings.TrimSuffix(rewrite, "/")
	newPath := strings.ReplaceAll(rewrite, "{path}", r.URL.Path)

	r.URL.Path = newPath
}
