package serverpool

import (
	"net/http"
	"sync"

	"github.com/letronghoangminh/reproxy/pkg/interfaces"
	"github.com/letronghoangminh/reproxy/pkg/utils"
)

type ipServerPool struct {
	backends []interfaces.Backend
	mux      sync.RWMutex
}

func (s *ipServerPool) GetNextValidPeer(r *http.Request) interfaces.Backend {
	s.mux.Lock()
	defer s.mux.Unlock()

	if len(s.backends) == 0 {
		return nil
	}

	return s.backends[utils.Hash(r.RemoteAddr)%uint32(len(s.backends))]
}

func (s *ipServerPool) AddBackend(b interfaces.Backend) {
	s.backends = append(s.backends, b)
}

func (s *ipServerPool) GetServerPoolSize() int {
	return len(s.backends)
}

func (s *ipServerPool) GetBackends() []interfaces.Backend {
	return s.backends
}
