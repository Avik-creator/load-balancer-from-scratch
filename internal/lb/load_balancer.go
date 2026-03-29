package lb

import (
	"errors"
	"log"
	"net/http"
	"sync/atomic"
)

// LoadBalancer is responsible for selecting a healthy backend
// and forwarding incoming requests to it.
//
// Selection strategy used here:
// - round-robin among healthy backends
//
// It is safe for concurrent use because the counter is atomic.
type LoadBalancer struct {
	backends []*Backend
	counter  atomic.Uint64
}

// New creates a new load balancer.
//
// It returns an error if no backends are provided because a load balancer
// without backends is useless.

func New(backends []*Backend) (*LoadBalancer, error) {
	if len(backends) == 0 {
		return nil, errors.New("at least one backend is required")
	}

	return &LoadBalancer{
		backends: backends,
	}, nil
}

// nextBackend selects the next healthy backend using round-robin.
//
// How it works:
// - pick a starting index using an atomic counter
// - scan all backends once
// - return the first healthy backend found
//
// If none are healthy, return an error instead of looping forever.

func (lb *LoadBalancer) nextBackend() (*Backend, error) {
	total := len(lb.backends)
	start := int(lb.counter.Add(1)-1) % total

	for i := 0; i < total; i++ {
		index := (start + i) % total
		backend := lb.backends[index]

		if backend.IsAlive() {
			return backend, nil
		}
	}

	return nil, errors.New("no healthy backends available")
}

// ServeHTTP makes LoadBalancer implement http.Handler.
//
// That lets us pass it directly to http.Server or http.ListenAndServe.
func (lb *LoadBalancer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	backend, err := lb.nextBackend()
	if err != nil {
		http.Error(rw, "no healthy backends available", http.StatusServiceUnavailable)
		return
	}

	log.Printf("forwarding %s %s to %s", req.Method, req.URL.Path, backend.URL)
	backend.Proxy.ServeHTTP(rw, req)
}
