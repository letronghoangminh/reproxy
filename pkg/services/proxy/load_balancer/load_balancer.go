package loadbalancer

import (
	"net/http"

	"github.com/letronghoangminh/reproxy/pkg/interfaces"
)

type RetryCount string

const (
	RETRY_COUNT RetryCount = "retry_count"
)

func AllowRetry(r *http.Request, maxRetries int) bool {
	if retryCount := r.Context().Value(RETRY_COUNT); retryCount != nil {
		if count, ok := retryCount.(int); ok {
			return count < maxRetries
		}
	}
	return true
}

type loadBalancer struct {
	serverPool interfaces.ServerPool
}

func (lb *loadBalancer) Serve(w http.ResponseWriter, r *http.Request) {
	peer := lb.serverPool.GetNextValidPeer(r)
	if peer != nil {
		peer.Serve(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func NewLoadBalancer(serverPool interfaces.ServerPool) interfaces.LoadBalancer {
	return &loadBalancer{
		serverPool: serverPool,
	}
}
