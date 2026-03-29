package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Avik-creator/load-balancer-from-scratch/internal/lb"
)

func main() {
	// Define upstream backends.
	// These should be your local test servers during development.
	rawBackends := []string{
		"http://localhost:9001",
		"http://localhost:9002",
		"http://localhost:9003",
	}

	var backends []*lb.Backend
	for _, raw := range rawBackends {
		backend, err := lb.NewBackend(raw)
		if err != nil {
			log.Fatalf("failed to create backend %s: %v", raw, err)
		}
		backends = append(backends, backend)
	}

	loadBalancer, err := lb.New(backends)
	if err != nil {
		log.Fatalf("failed to create load balancer: %v", err)
	}

	// Start background health checks.
	lb.StartHealthChecks(backends, 5*time.Second)

	// Create a proper HTTP server instead of blindly using ListenAndServe.
	// This gives cleaner configuration and is easier to extend later.
	server := &http.Server{
		Addr:              ":8000",
		Handler:           loadBalancer,
		ReadHeaderTimeout: 3 * time.Second,
	}

	log.Println("load balancer listening on http://localhost:8000")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
