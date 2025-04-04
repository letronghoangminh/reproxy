package serverpool

import (
	"fmt"
	"sync"

	"github.com/letronghoangminh/reproxy/pkg/services/reverse_proxy/backend"
)

type ServerPool interface {
	GetBackends() []backend.Backend
	GetNextValidPeer() backend.Backend
	AddBackend(backend.Backend)
	GetServerPoolSize() int
}

type roundRobinServerPool struct {
	backends []backend.Backend
	mux      sync.RWMutex
	current  int
}

func (s *roundRobinServerPool) Rotate() backend.Backend {
	s.mux.Lock()
	s.current = (s.current + 1) % s.GetServerPoolSize()
	s.mux.Unlock()
	return s.backends[s.current]
}

func (s *roundRobinServerPool) GetNextValidPeer() backend.Backend {
	for i := 0; i < s.GetServerPoolSize(); i++ {
		nextPeer := s.Rotate()
		if nextPeer.IsAlive() {
			return nextPeer
		}
	}
	return nil
}

func (s *roundRobinServerPool) GetBackends() []backend.Backend {
	return s.backends
}

func (s *roundRobinServerPool) AddBackend(b backend.Backend) {
	s.backends = append(s.backends, b)
}

func (s *roundRobinServerPool) GetServerPoolSize() int {
	return len(s.backends)
}

func NewServerPool(strategy LBStrategy) (ServerPool, error) {
	switch strategy {
	case RoundRobin:
		return &roundRobinServerPool{
			backends: make([]backend.Backend, 0),
			current:  0,
		}, nil
	case LeastConnections:
		return &lcServerPool{
			backends: make([]backend.Backend, 0),
		}, nil
	case Random:
		return &randomServerPool{
			backends: make([]backend.Backend, 0),
		}, nil
	// case IPHash:
	// 	return &ipHashServerPool{
	// 		backends: make([]backend.Backend, 0),
	// 	}, nil
	// case URIHash:
	// 	return &uriHashServerPool{
	// 		backends: make([]backend.Backend, 0),
	// 	}, nil
	default:
		return nil, fmt.Errorf("Invalid strategy")
	}
}
