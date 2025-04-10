package serverpool

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/letronghoangminh/reproxy/pkg/interfaces"
)

type stickyServerPool struct {
	backends []interfaces.Backend
	mux      sync.RWMutex
	current  int
}

func (s *stickyServerPool) Rotate() interfaces.Backend {
	s.mux.Lock()
	s.current = (s.current + 1) % s.GetServerPoolSize()
	s.mux.Unlock()
	return s.backends[s.current]
}

func (s *stickyServerPool) GetNextValidPeer(r *http.Request) interfaces.Backend {
	stickyCookie, err := r.Cookie("X-Sticky-Session-ID")
	if err == nil {
		stickySessionID, err := strconv.Atoi(stickyCookie.Value)

		if err == nil {
			for idx, b := range s.backends {
				if b.IsAlive() && stickySessionID == idx {
					return b
				}
			}
		}
	}

	for i := 0; i < s.GetServerPoolSize(); i++ {
		nextPeer := s.Rotate()
		if nextPeer.IsAlive() {
			cookie := &http.Cookie{
				Name:  "X-Sticky-Session-ID",
				Value: strconv.Itoa(s.current),
			}
			nextPeer.AddCookie(cookie)
			return nextPeer
		}
	}

	return nil
}

func (s *stickyServerPool) GetBackends() []interfaces.Backend {
	return s.backends
}

func (s *stickyServerPool) AddBackend(b interfaces.Backend) {
	s.backends = append(s.backends, b)
}

func (s *stickyServerPool) GetServerPoolSize() int {
	return len(s.backends)
}
