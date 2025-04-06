package serverpool

import (
	"net/http"
	"sync"

	"github.com/letronghoangminh/reproxy/pkg/services/reverse_proxy/backend"
)

type lcServerPool struct {
	backends []backend.Backend
	mux      sync.RWMutex
}

func (s *lcServerPool) GetNextValidPeer(r *http.Request) backend.Backend {
	var leastConnectedPeer backend.Backend
	for _, b := range s.backends {
		if b.IsAlive() {
			leastConnectedPeer = b
			break
		}
	}

	for _, b := range s.backends {
		if !b.IsAlive() {
			continue
		}
		if leastConnectedPeer.GetActiveConnections() > b.GetActiveConnections() {
			leastConnectedPeer = b
		}
	}
	return leastConnectedPeer
}

func (s *lcServerPool) AddBackend(b backend.Backend) {
	s.backends = append(s.backends, b)
}

func (s *lcServerPool) GetServerPoolSize() int {
	return len(s.backends)
}

func (s *lcServerPool) GetBackends() []backend.Backend {
	return s.backends
}
