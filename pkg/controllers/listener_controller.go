package controllers

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/letronghoangminh/reproxy/pkg/config"
	"github.com/letronghoangminh/reproxy/pkg/services"
	"github.com/letronghoangminh/reproxy/pkg/utils"
	"go.uber.org/zap"
)

type ListenerController struct {
	Server        *http.ServeMux
	Port          int
	TargetHandler map[string][]*config.HandlerConfig
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
			utils.Logger.Info("initializing new listener controller", zap.Int("port", port))
			listenerControllers[port] = ListenerController{
				Server:        http.NewServeMux(),
				Port:          port,
				TargetHandler: map[string][]*config.HandlerConfig{},
			}
			listenerControllers[port].Server.HandleFunc("/", defaultHandler)
		}

		handlerPointers := make([]*config.HandlerConfig, len(handlers))
		for i := range handlers {
			handlerPointers[i] = &handlers[i]
			if handlers[i].ReverseProxy.Upstreams != nil {
				reverseProxyHandlers = append(reverseProxyHandlers, &handlers[i])
			}
		}
		listenerControllers[port].TargetHandler[hostname] = handlerPointers
	}

	services.StartLoadBalancers(ctx, reverseProxyHandlers)

	for port, listenerController := range listenerControllers {
		port := port
		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: listenerController.Server,
		}
		
		wg.Add(1)
		go func() {
			utils.Logger.Info("serving new controller", zap.Int("port", port))
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				utils.Logger.Error(fmt.Sprintf("error occurred while serving controller on port %d", port), zap.Error(err))
			}
		}()
		
		go func() {
			<-ctx.Done()
			utils.Logger.Info("shutting down controller", zap.Int("port", port))
			if err := server.Shutdown(context.Background()); err != nil {
				utils.Logger.Error(fmt.Sprintf("error shutting down controller on port %d", port), zap.Error(err))
			}
			wg.Done()
		}()
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
	utils.Logger.Info("requesting coming", zap.String("path", r.URL.Path), zap.String("host", r.Host))

	var port int; var host string

	if strings.Contains(r.Host, ":") {
		port, _ = strconv.Atoi(strings.Split(r.Host, ":")[1])
		host = strings.Split(r.Host, ":")[0]
	} else {
		host = r.Host
	}

	listenerController := listenerControllers[port]
	handlers, ok := listenerController.TargetHandler[host]

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 page not found"))
		return
	}

	for _, handler := range handlers {
		if strings.HasPrefix(r.URL.Path, handler.Path) {
			switch {
			case handler.StaticResponse.Body != "":
				w.WriteHeader(handler.StaticResponse.StatusCode)
				w.Write([]byte(handler.StaticResponse.Body))
				return
			case handler.StaticFiles.Root != "":
				cleanPath := path.Clean(strings.TrimPrefix(r.URL.Path, handler.Path))
				for strings.HasPrefix(cleanPath, "..") || strings.HasPrefix(cleanPath, "/") {
					cleanPath = strings.TrimPrefix(cleanPath, "..")
					cleanPath = strings.TrimPrefix(cleanPath, "/")
				}
				filePath := path.Join(handler.StaticFiles.Root, cleanPath)
				utils.Logger.Debug("serving static file", zap.String("filePath", filePath))
				http.ServeFile(w, r, filePath)
				return
			case handler.ReverseProxy.Upstreams != nil:
				utils.Logger.Debug("serving reverse proxy", zap.Strings("upstream", handler.ReverseProxy.Upstreams))
				services.HandleReverseProxyRequest(w, r, handler)
				return
			}
		}
	}

	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 page not found"))
}
