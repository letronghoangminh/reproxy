package serverpool

import (
	"net/http"
	"sync"

	"github.com/letronghoangminh/reproxy/pkg/interfaces"
)

type roundRobinServerPool struct {
	backends []interfaces.Backend
	mux      sync.RWMutex
	current  int
}

func (s *roundRobinServerPool) Rotate() interfaces.Backend {
	s.mux.Lock()
	s.current = (s.current + 1) % s.GetServerPoolSize()
	s.mux.Unlock()
	return s.backends[s.current]
}

func (s *roundRobinServerPool) GetNextValidPeer(r *http.Request) interfaces.Backend {
	for i := 0; i < s.GetServerPoolSize(); i++ {
		nextPeer := s.Rotate()
		if nextPeer.IsAlive() {
			return nextPeer
		}
	}
	return nil
}

func (s *roundRobinServerPool) GetBackends() []interfaces.Backend {
	return s.backends
}

func (s *roundRobinServerPool) AddBackend(b interfaces.Backend) {
	s.backends = append(s.backends, b)
}

func (s *roundRobinServerPool) GetServerPoolSize() int {
	return len(s.backends)
}
