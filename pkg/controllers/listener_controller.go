// Package controllers provides functionality to manage HTTP listeners and route requests to appropriate handlers.
package controllers

import (
	"compress/gzip"
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/letronghoangminh/reproxy/pkg/config"
	"github.com/letronghoangminh/reproxy/pkg/services/matcher"
	"github.com/letronghoangminh/reproxy/pkg/services/proxy"
	"github.com/letronghoangminh/reproxy/pkg/services/static"
	"github.com/letronghoangminh/reproxy/pkg/utils"
)

type ListenerController struct {
	Server        *http.ServeMux
	Port          int
	TargetHandler map[string][]*config.HandlerConfig
}

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (gzw *gzipResponseWriter) Write(b []byte) (int, error) {
	return gzw.Writer.Write(b)
}

var (
	listenerControllers map[int]ListenerController
)

func InitListenerControllers(ctx context.Context, wg *sync.WaitGroup) {
	listenerControllers = map[int]ListenerController{}
	reverseProxyHandlers := []*config.HandlerConfig{}

	utils.Logger.Info("parsing listener configs")
	listeners := combineListener()

	for host, handlers := range listeners {
		utils.Logger.Info("constructing listener controllers")

		port, _ := strconv.Atoi(strings.Split(host, ":")[1])
		hostname := strings.Split(host, ":")[0]

		_, ok := listenerControllers[port]
		if !ok {
			utils.Logger.Info("initializing new listener controller", "port", port)
			listenerControllers[port] = ListenerController{
				Server:        http.NewServeMux(),
				Port:          port,
				TargetHandler: map[string][]*config.HandlerConfig{},
			}
			listenerControllers[port].Server.HandleFunc("/", gzipHandler(defaultHandler))
		}

		handlerPointers := make([]*config.HandlerConfig, len(handlers))
		for i := range handlers {
			handlerPointers[i] = &handlers[i]
			if len(handlers[i].ReverseProxy.Upstreams.Dynamic) > 0 || len(handlers[i].ReverseProxy.Upstreams.Static) > 0 {
				reverseProxyHandlers = append(reverseProxyHandlers, &handlers[i])
			}
		}
		listenerControllers[port].TargetHandler[hostname] = handlerPointers
	}

	proxy.StartLoadBalancers(ctx, reverseProxyHandlers)

	for port, listenerController := range listenerControllers {
		port := port
		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: listenerController.Server,
		}

		wg.Add(1)
		go func() {
			utils.Logger.Info("serving new controller", "port", port)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				utils.Logger.Error(fmt.Sprintf("error occurred while serving controller on port %d", port), "error", err)
			}
		}()

		go func() {
			<-ctx.Done()
			utils.Logger.Info("shutting down controller", "port", port)
			if err := server.Shutdown(context.Background()); err != nil {
				utils.Logger.Error(fmt.Sprintf("error shutting down controller on port %d", port), "error", err)
			}
			wg.Done()
		}()
	}
}

func gzipHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			next(w, r)
			return
		}
		defer gz.Close()

		gzw := &gzipResponseWriter{
			ResponseWriter: w,
			Writer:         gz,
		}

		w.Header().Set("Content-Encoding", "gzip")

		next(gzw, r)
	}
}

func combineListener() map[string][]config.HandlerConfig {
	listeners := map[string][]config.HandlerConfig{}

	for _, listenerConfig := range cfg.Listeners {
		for _, host := range listenerConfig.Host {
			handlers := listenerConfig.Handlers

			_, ok := listeners[host]
			if !ok {
				listeners[host] = handlers
				continue
			}

			listeners[host] = append(listeners[host], handlers...)
		}
	}

	return listeners
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()
	logger.Info("Request received",
		"path", r.URL.Path,
		"host", r.Host,
		"method", r.Method,
		"remote_addr", r.RemoteAddr)

	// Add request ID for tracing
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = utils.GenerateRequestID()
		r.Header.Set("X-Request-ID", requestID)
	}
	logger = logger.With("request_id", requestID)

	var port int
	var host string

	if strings.Contains(r.Host, ":") {
		var err error
		var portStr string
		host, portStr, err = net.SplitHostPort(r.Host)
		if err != nil {
			logger.Error("Failed to parse host and port", "host", r.Host, "error", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		port, err = strconv.Atoi(portStr)
		if err != nil {
			logger.Error("Invalid port number", "port", portStr, "error", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
	} else {
		host = r.Host
		port = cfg.Global.Port
	}

	listenerController, ok := listenerControllers[port]
	if !ok {
		logger.Warn("No listener controller for port", "port", port)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	handlers, ok := listenerController.TargetHandler[host]
	if !ok {
		logger.Warn("No handlers for host", "host", host)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	handler := matcher.MatchHandler(r, handlers)
	if handler != nil {
		handleRequest(w, r, handler)
	} else {
		logger.Debug("No matching handler found")
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request, handler *config.HandlerConfig) {
	logger := utils.GetLogger()

	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = utils.GenerateRequestID()
		r.Header.Set("X-Request-ID", requestID)
	}
	logger = logger.With("request_id", requestID)

	switch {
	case handler.StaticResponse.Body != "":
		logger.Debug("Handling static response")
		err := static.ServeStaticResponse(w, r, handler)
		if err != nil {
			logger.Error("Failed to serve static response", "error", err)
		}

	case handler.StaticFiles.Root != "":
		logger.Debug("Handling static file")
		err := static.ServeFile(w, r, handler)
		if err != nil {
			logger.Error("Failed to serve static file", "error", err)
		}

	case len(handler.ReverseProxy.Upstreams.Dynamic) > 0 || len(handler.ReverseProxy.Upstreams.Static) > 0:
		logger.Debug("Handling reverse proxy")
		proxy.HandleReverseProxyRequest(w, r, handler)

	default:
		logger.Warn("Handler matched but no implementation found",
			"matcher_path", handler.Matchers.Path)
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}
}
