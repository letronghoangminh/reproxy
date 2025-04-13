package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/letronghoangminh/reproxy/pkg/config"
	"github.com/letronghoangminh/reproxy/pkg/utils"
)

var (
	cfg *config.Config
)

func retrieveConfig(w http.ResponseWriter, r *http.Request) {
	utils.Logger.Info("requesting for retrieving server config", "method", r.Method, "path", r.URL.Path)

	jsonCfg, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		utils.Logger.Error("error occurred while marshalling config", "error", err)
		http.Error(w, "error occurred while marshalling config", http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(jsonCfg)
}

func DefaultControllerServe(ctx context.Context, wg *sync.WaitGroup) {
	cfg = config.GetConfig()
	port := cfg.Global.Port

	http.HandleFunc("/config", retrieveConfig)

	utils.Logger.Info("default controller is serving", "port", port)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: nil,
	}

	wg.Add(1)
	go func() {
		utils.Logger.Info("serving default controller", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Logger.Error(fmt.Sprintf("error occurred while serving default controller on port %d", port), "error", err)
		}
	}()

	go func() {
		<-ctx.Done()
		utils.Logger.Info("shutting down default controller", "port", port)
		if err := server.Shutdown(context.Background()); err != nil {
			utils.Logger.Error(fmt.Sprintf("error shutting down default controller on port %d", port), "error", err)
		}
		wg.Done()
	}()
}
