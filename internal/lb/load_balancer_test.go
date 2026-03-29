package lb

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// helperBackend creates a backend backed by an httptest server.
// This is useful for unit tests because it gives us real HTTP servers
// without manual setup.
func helperBackend(t *testing.T, handler http.HandlerFunc) *Backend {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	backend, err := NewBackend(server.URL)
	if err != nil {
		t.Fatalf("failed to create backend: %v", err)
	}

	backend.SetAlive(true)
	return backend
}

func TestNextBackend_RoundRobin(t *testing.T) {
	b1 := helperBackend(t, func(w http.ResponseWriter, r *http.Request) {})
	b2 := helperBackend(t, func(w http.ResponseWriter, r *http.Request) {})
	b3 := helperBackend(t, func(w http.ResponseWriter, r *http.Request) {})

	lb, err := New([]*Backend{b1, b2, b3})
	if err != nil {
		t.Fatalf("failed to create load balancer: %v", err)
	}

	got1, err := lb.nextBackend()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got2, err := lb.nextBackend()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got3, err := lb.nextBackend()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got1 != b1 {
		t.Fatalf("expected b1, got %v", got1.URL)
	}
	if got2 != b2 {
		t.Fatalf("expected b2, got %v", got2.URL)
	}
	if got3 != b3 {
		t.Fatalf("expected b3, got %v", got3.URL)
	}
}

func TestNextBackend_SkipsDeadBackend(t *testing.T) {
	b1 := helperBackend(t, func(w http.ResponseWriter, r *http.Request) {})
	b2 := helperBackend(t, func(w http.ResponseWriter, r *http.Request) {})
	b3 := helperBackend(t, func(w http.ResponseWriter, r *http.Request) {})

	b2.SetAlive(false)

	lb, err := New([]*Backend{b1, b2, b3})
	if err != nil {
		t.Fatalf("failed to create load balancer: %v", err)
	}

	got1, err := lb.nextBackend()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got2, err := lb.nextBackend()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got3, err := lb.nextBackend()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got1 != b1 {
		t.Fatalf("expected b1, got %v", got1.URL)
	}
	if got2 != b3 {
		t.Fatalf("expected b3, got %v", got2.URL)
	}
	if got3 != b1 {
		t.Fatalf("expected b1 again, got %v", got3.URL)
	}
}

func TestNextBackend_AllDead(t *testing.T) {
	b1 := helperBackend(t, func(w http.ResponseWriter, r *http.Request) {})
	b2 := helperBackend(t, func(w http.ResponseWriter, r *http.Request) {})

	b1.SetAlive(false)
	b2.SetAlive(false)

	lb, err := New([]*Backend{b1, b2})
	if err != nil {
		t.Fatalf("failed to create load balancer: %v", err)
	}

	_, err = lb.nextBackend()
	if err == nil {
		t.Fatal("expected error when all backends are dead, got nil")
	}
}

func TestServeHTTP_ProxiesToBackend(t *testing.T) {
	b1 := helperBackend(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("backend-1"))
	})

	lb, err := New([]*Backend{b1})
	if err != nil {
		t.Fatalf("failed to create load balancer: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "http://localhost:8000/", nil)
	rec := httptest.NewRecorder()

	lb.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	if rec.Body.String() != "backend-1" {
		t.Fatalf("expected body backend-1, got %q", rec.Body.String())
	}
}

func TestServeHTTP_Returns503WhenNoBackendsAlive(t *testing.T) {
	b1 := helperBackend(t, func(w http.ResponseWriter, r *http.Request) {})
	b1.SetAlive(false)

	lb, err := New([]*Backend{b1})
	if err != nil {
		t.Fatalf("failed to create load balancer: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "http://localhost:8000/", nil)
	rec := httptest.NewRecorder()

	lb.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}
