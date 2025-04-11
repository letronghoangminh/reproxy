package serverpool

import (
	"net/http"
	"sync"

	"github.com/letronghoangminh/reproxy/pkg/interfaces"
	"github.com/letronghoangminh/reproxy/pkg/utils"
)

type uriServerPool struct {
	backends []interfaces.Backend
	mux      sync.RWMutex
}

func (s *uriServerPool) GetNextValidPeer(r *http.Request) interfaces.Backend {
	s.mux.Lock()
	defer s.mux.Unlock()

	if len(s.backends) == 0 {
		return nil
	}

	return s.backends[utils.Hash(r.URL.Path)%uint32(len(s.backends))]
}

func (s *uriServerPool) AddBackend(b interfaces.Backend) {
	s.backends = append(s.backends, b)
}

func (s *uriServerPool) GetServerPoolSize() int {
	return len(s.backends)
}

func (s *uriServerPool) GetBackends() []interfaces.Backend {
	return s.backends
}
