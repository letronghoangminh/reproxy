package interfaces

import (
	"net/http"
)

type ServerPool interface {
	GetBackends() []Backend

	GetNextValidPeer(r *http.Request) Backend

	AddBackend(Backend)

	GetServerPoolSize() int
}
