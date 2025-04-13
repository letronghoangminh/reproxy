// Package loadbalancer provides functionality to handle load balancing for reverse proxy requests.
package loadbalancer

import (
	"net/http"

	"github.com/letronghoangminh/reproxy/pkg/interfaces"
)

type RetryCount string

const (
	Count RetryCount = "retry_count"
)

func AllowRetry(r *http.Request, maxRetries int) bool {
	if retryCount := r.Context().Value(Count); retryCount != nil {
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
