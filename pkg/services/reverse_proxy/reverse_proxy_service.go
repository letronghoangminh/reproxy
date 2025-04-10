package reverseproxy

import (
	"context"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/letronghoangminh/reproxy/pkg/config"
	"github.com/letronghoangminh/reproxy/pkg/services/dns"
	"github.com/letronghoangminh/reproxy/pkg/services/reverse_proxy/backend"
	loadbalancer "github.com/letronghoangminh/reproxy/pkg/services/reverse_proxy/load_balancer"
	"github.com/letronghoangminh/reproxy/pkg/services/reverse_proxy/serverpool"
	"github.com/letronghoangminh/reproxy/pkg/utils"
	"go.uber.org/zap"
)

// Reverse proxy implementation credits to https://github.com/leonardo5621/golang-load-balancer

var (
	loadBalancers = map[*config.HandlerConfig]loadbalancer.LoadBalancer{}
)

func StartLoadBalancers(ctx context.Context, handlers []*config.HandlerConfig) {
	for _, handler := range handlers {
		serverPool, err := serverpool.NewServerPool(serverpool.GetLBStrategy(handler.ReverseProxy.LoadBalancing.Strategy))
		if err != nil {
			utils.Logger.Error("error occurred while creating server pool", zap.Error(err))
			return
		}

		loadBalancer := loadbalancer.NewLoadBalancer(serverPool)

		staticUpstreams := handler.ReverseProxy.Upstreams.Static
		dynamicUpstreams := dns.GetDynamicUpstreams(handler.ReverseProxy.Upstreams.Dynamic)

		for _, u := range append(staticUpstreams, dynamicUpstreams...) {
			endpoint, err := url.Parse(u)
			if err != nil {
				utils.Logger.Fatal(err.Error(), zap.String("URL", u))
			}

			rp := httputil.NewSingleHostReverseProxy(endpoint)

			backendServer := backend.NewBackend(endpoint, rp)
			rp.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
				utils.Logger.Error("error handling the request",
					zap.String("host", endpoint.Host),
					zap.Error(e),
				)
				backendServer.SetAlive(false)

				if !loadbalancer.AllowRetry(request, handler.ReverseProxy.LoadBalancing.Retries) {
					utils.Logger.Info(
						"Max retry attempts reached, terminating",
						zap.String("address", request.RemoteAddr),
						zap.String("path", request.URL.Path),
					)
					http.Error(writer, "Service not available", http.StatusServiceUnavailable)
					return
				}

				currentCount, ok := request.Context().Value(loadbalancer.RETRY_COUNT).(int)
				if !ok {
					currentCount = 0
				}

				utils.Logger.Info(
					"Attempting retry",
					zap.String("address", request.RemoteAddr),
					zap.String("URL", request.URL.Path),
					zap.Bool("retry", true),
					zap.Int("retry_count", currentCount+1),
				)

				sleepDuration := handler.ReverseProxy.LoadBalancing.TryInterval
				if sleepDuration == 0 {
					sleepDuration = 5
				}
				time.Sleep(time.Duration(handler.ReverseProxy.LoadBalancing.TryInterval) * time.Second)

				loadBalancer.Serve(
					writer,
					request.WithContext(
						context.WithValue(request.Context(), loadbalancer.RETRY_COUNT, currentCount+1),
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

	if handler.Path != "" {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, handler.Path)
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
		result = strings.Replace(result, placeholder, replacement, -1)
	}

	return result
}

func rewritePath(r *http.Request, rewrite string) {
	if rewrite == "" {
		return
	}

	rewrite = strings.TrimPrefix(rewrite, "/")
	rewrite = strings.TrimSuffix(rewrite, "/")
	newPath := strings.Replace(rewrite, "{path}", r.URL.Path, -1)

	r.URL.Path = newPath
}
