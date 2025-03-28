package controllers

import (
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/letronghoangminh/reproxy/pkg/config"
	"github.com/letronghoangminh/reproxy/pkg/models"
	"go.uber.org/zap"
)

var (
	listenerControllers map[int]models.ListenerController
)

func InitListenerControllers() {
	logger = zap.L()
	listenerControllers = map[int]models.ListenerController{}

	logger.Info("parsing listener configs")
	listeners := combineListener()

	for host, handlers  := range listeners {
		logger.Info("constructing listener controllers")

		port, _ := strconv.Atoi(strings.Split(host, ":")[1])
		hostname := strings.Split(host, ":")[0]

		_, ok := listenerControllers[port]
		if !ok {
			logger.Info("initializing new listener controller", zap.Int("port", port))
			listenerControllers[port] = models.ListenerController{
				Server:        http.NewServeMux(),
				Port:          port,
				TargetHandler: map[string][]config.HandlerConfig{},
			}
			listenerControllers[port].Server.HandleFunc("/", defaultHandler)
		}

		listenerControllers[port].TargetHandler[hostname] = handlers
	}

	for port, listenerController := range listenerControllers {
		go func() {
			logger.Info("serving new controller", zap.Int("port", port))
			err := http.ListenAndServe(fmt.Sprintf(":%d", port), listenerController.Server)
			if err != nil {
				logger.Error(fmt.Sprintf("error occurred while serving new controller on port %d", port), zap.Error(err))
			}
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
	logger.Info("requesting coming", zap.String("path", r.URL.Path), zap.String("host", r.Host))

	port, _ := strconv.Atoi(strings.Split(r.Host, ":")[1])
	host := strings.Split(r.Host, ":")[0]

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
				logger.Debug("serving static file", zap.String("filePath", filePath))
				http.ServeFile(w, r, filePath)
				return
			case handler.ReverseProxy.Upstreams != nil:
				return
			}
		}
	}

	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 page not found"))
}
