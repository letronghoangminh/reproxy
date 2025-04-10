package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/letronghoangminh/reproxy/pkg/config"
	"github.com/letronghoangminh/reproxy/pkg/utils"
	"go.uber.org/zap"
)

var (
	cfg *config.Config
)

func retrieveConfig(w http.ResponseWriter, r *http.Request) {
	utils.Logger.Info("requesting for retrieving server config", zap.String("method", r.Method), zap.String("path", r.URL.Path))

	jsonCfg, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		utils.Logger.Error("error occurred while marshalling config", zap.Error(err))
		http.Error(w, "error occurred while marshalling config", http.StatusInternalServerError)
		return
	}

	w.Write(jsonCfg)
}

func DefaultControllerServe(ctx context.Context, wg *sync.WaitGroup) {
	cfg = config.GetConfig()
	port := cfg.Global.Port

	http.HandleFunc("/config", retrieveConfig)

	utils.Logger.Info("default controller is serving", zap.Int("port", port))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: nil,
	}

	wg.Add(1)
	go func() {
		utils.Logger.Info("serving default controller", zap.Int("port", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Logger.Error(fmt.Sprintf("error occurred while serving default controller on port %d", port), zap.Error(err))
		}
	}()

	go func() {
		<-ctx.Done()
		utils.Logger.Info("shutting down default controller", zap.Int("port", port))
		if err := server.Shutdown(context.Background()); err != nil {
			utils.Logger.Error(fmt.Sprintf("error shutting down default controller on port %d", port), zap.Error(err))
		}
		wg.Done()
	}()
}
