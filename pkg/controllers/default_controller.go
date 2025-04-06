package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/letronghoangminh/reproxy/pkg/config"
	"go.uber.org/zap"
)

var (
	logger *zap.Logger
	cfg    *config.Config
)

func retrieveConfig(w http.ResponseWriter, r *http.Request) {
	logger.Info("requesting for retrieving server config", zap.String("method", r.Method), zap.String("path", r.URL.Path))

	jsonCfg, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		logger.Error("error occurred while marshalling config", zap.Error(err))
		http.Error(w, "error occurred while marshalling config", http.StatusInternalServerError)
		return
	}

	w.Write(jsonCfg)
}

func DefaultControllerServe(ctx context.Context, wg *sync.WaitGroup) {
	cfg = config.GetConfig()
	logger = zap.L()
	port := cfg.Global.Port

	http.HandleFunc("/config", retrieveConfig)

	logger.Info("default controller is serving", zap.Int("port", port))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: nil,
	}

	wg.Add(1)
	go func() {
		logger.Info("serving default controller", zap.Int("port", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(fmt.Sprintf("error occurred while serving default controller on port %d", port), zap.Error(err))
		}
	}()
	
	go func() {
		<-ctx.Done()
		logger.Info("shutting down default controller", zap.Int("port", port))
		if err := server.Shutdown(context.Background()); err != nil {
			logger.Error(fmt.Sprintf("error shutting down default controller on port %d", port), zap.Error(err))
		}
		wg.Done()
	}()
}
