package lb

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"
)

// Backend represents one upstream server behind the load balancer.
//
// It stores:
// - the parsed upstream URL
// - a reverse proxy configured to forward traffic to that upstream
// - an atomic alive flag so health state can be safely read/written
//   across many goroutines

type Backend struct {
	URL   *url.URL
	Alive atomic.Bool
	Proxy *httputil.ReverseProxy
}

// NewBackend creates a backend from a raw URL string.
//
// Example:
//   b, err := NewBackend("http://localhost:9001")
//
// We configure a custom reverse proxy transport so the proxy behaves
// more predictably under timeouts and reused connections.

func NewBackend(rawURL string) (*Backend, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid backend URL: %w", err)

	}

	backend := &Backend{
		URL: parsedURL,
	}
	backend.Alive.Store(true)

	proxy := httputil.NewSingleHostReverseProxy(parsedURL)
	// Custom transport gives us better control than the default one.
	// These settings are reasonable for a learning project and much
	// better than relying blindly on defaults.
	proxy.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   3 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
	}

	// ErrorHandler is called when the reverse proxy fails to talk to
	// the upstream server.
	//
	// We mark the backend unhealthy so future requests can avoid it
	// until the health checker marks it alive again.
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		log.Printf("proxy error for backend %s: %v", backend.URL, err)
		backend.SetAlive(false)
		http.Error(rw, "upstream server unavailable", http.StatusBadGateway)
	}

	backend.Proxy = proxy
	return backend, nil

}

// IsAlive returns whether the backend is currently considered healthy.
func (b *Backend) IsAlive() bool {
	return b.Alive.Load()
}

// SetAlive updates the backend health flag.
func (b *Backend) SetAlive(alive bool) {
	b.Alive.Store(alive)
}
