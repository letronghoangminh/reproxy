package serverpool

import (
	"math/rand"
	"net/http"
	"sync"

	"github.com/letronghoangminh/reproxy/pkg/interfaces"
)

type randomServerPool struct {
	backends []interfaces.Backend
	mux      sync.RWMutex
}

func (s *randomServerPool) GetNextValidPeer(r *http.Request) interfaces.Backend {
	s.mux.Lock()
	defer s.mux.Unlock()

	if len(s.backends) == 0 {
		return nil
	}

	randomIndex := rand.Intn(len(s.backends))
	return s.backends[randomIndex]
}

func (s *randomServerPool) AddBackend(b interfaces.Backend) {
	s.backends = append(s.backends, b)
}

func (s *randomServerPool) GetServerPoolSize() int {
	return len(s.backends)
}

func (s *randomServerPool) GetBackends() []interfaces.Backend {
	return s.backends
}
