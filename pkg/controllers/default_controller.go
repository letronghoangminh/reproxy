package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/letronghoangminh/reproxy/pkg/config"
	"go.uber.org/zap"
)

var (
	logger *zap.Logger
	cfg *config.Config
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

func DefaultControllerServe() {
	cfg = config.GetConfig()
	logger = zap.L()

	http.HandleFunc("/config", retrieveConfig)

	logger.Info("default controller is serving", zap.Int("port", cfg.Global.Port))

	err := http.ListenAndServe(fmt.Sprintf(":%v", cfg.Global.Port), nil)
	if err != nil {
		logger.Error("error occurred while serving default controller", zap.Error(err))
	}
}
