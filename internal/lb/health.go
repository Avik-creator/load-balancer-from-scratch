package lb

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

// CheckBackendHealth performs one active health check against:
//
//	<backend-url>/health
//
// A backend is considered healthy only if the response returns HTTP 200.
func CheckBackendHealth(ctx context.Context, backend *Backend) {
	healthURL := fmt.Sprintf("%s/health", backend.URL.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		backend.SetAlive(false)
		return
	}

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		backend.SetAlive(false)
		return
	}
	defer resp.Body.Close()

	backend.SetAlive(resp.StatusCode == http.StatusOK)
}

// StartHealthChecks launches a goroutine that periodically checks
// all backends.
//
// This is active health checking.
// If a backend goes down, it gets marked unhealthy.
// If it comes back and starts returning 200 on /health, it gets marked healthy again.
func StartHealthChecks(backends []*Backend, interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()

		for range ticker.C {
			for _, backend := range backends {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				CheckBackendHealth(ctx, backend)
				cancel()

				log.Printf("health check: backend=%s alive=%v", backend.URL, backend.IsAlive())
			}
		}
	}()
}
