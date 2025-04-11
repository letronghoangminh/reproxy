package interfaces

import (
	"net/http"

	"github.com/letronghoangminh/reproxy/pkg/config"
)

type Matcher interface {
	MatchHandler(r *http.Request, handlers []*config.HandlerConfig) *config.HandlerConfig
}
