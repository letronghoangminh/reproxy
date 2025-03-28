package models

import (
	"net/http"

	"github.com/letronghoangminh/reproxy/pkg/config"
)

type ListenerController struct {
	Server *http.ServeMux
	Port int
	TargetHandler map[string][]config.HandlerConfig
}
