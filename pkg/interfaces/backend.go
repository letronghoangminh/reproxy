package interfaces

import (
	"net/http"
	"net/url"
)

type Backend interface {
	SetAlive(bool)

	IsAlive() bool

	GetURL() *url.URL

	GetActiveConnections() int

	Serve(http.ResponseWriter, *http.Request)

	AddCookie(*http.Cookie)
}
