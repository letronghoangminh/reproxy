package loadbalancer

import (
	"net/http"

	"github.com/letronghoangminh/reproxy/pkg/services/reverse_proxy/serverpool"
)

const (
	RETRY_COUNT string = "retry_count"
)

func AllowRetry(r *http.Request, maxRetries int) bool {
	if retryCount := r.Context().Value(RETRY_COUNT); retryCount != nil {
		if count, ok := retryCount.(int); ok {
			return count < maxRetries
		}
	}
	return true
}

type LoadBalancer interface {
	Serve(http.ResponseWriter, *http.Request)
}

type loadBalancer struct {
	serverPool serverpool.ServerPool
}

func (lb *loadBalancer) Serve(w http.ResponseWriter, r *http.Request) {
	peer := lb.serverPool.GetNextValidPeer(r)
	if peer != nil {
		peer.Serve(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func NewLoadBalancer(serverPool serverpool.ServerPool) LoadBalancer {
	return &loadBalancer{
		serverPool: serverPool,
	}
}
