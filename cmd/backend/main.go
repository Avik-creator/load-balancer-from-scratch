package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

// This binary acts as a tiny backend server for local testing.
//
// Example:
//
//	go run ./cmd/backend -port=9001 -name=backend-1
//
// Then run it again on other ports:
//
//	go run ./cmd/backend -port=9002 -name=backend-2
//	go run ./cmd/backend -port=9003 -name=backend-3
func main() {
	port := flag.String("port", "9001", "port to listen on")
	name := flag.String("name", "backend", "backend server name")
	flag.Parse()

	mux := http.NewServeMux()

	// Health endpoint used by the load balancer.
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Main handler so you can see which backend served the request.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "response from %s on port %s\n", *name, *port)
	})

	addr := ":" + *port
	log.Printf("%s listening on http://localhost%s", *name, addr)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}
