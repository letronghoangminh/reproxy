package serverpool

import (
	"net/http"
	"sync"

	"github.com/letronghoangminh/reproxy/pkg/services/reverse_proxy/backend"
	"github.com/letronghoangminh/reproxy/pkg/utils"
)

type ipServerPool struct {
	backends []backend.Backend
	mux      sync.RWMutex
}

func (s *ipServerPool) GetNextValidPeer(r *http.Request) backend.Backend {
	s.mux.Lock()
	defer s.mux.Unlock()

	if len(s.backends) == 0 {
		return nil
	}

	return s.backends[utils.Hash(r.RemoteAddr)%uint32(len(s.backends))]
}

func (s *ipServerPool) AddBackend(b backend.Backend) {
	s.backends = append(s.backends, b)
}

func (s *ipServerPool) GetServerPoolSize() int {
	return len(s.backends)
}

func (s *ipServerPool) GetBackends() []backend.Backend {
	return s.backends
}
