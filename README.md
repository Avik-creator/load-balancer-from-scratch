# Go Load Balancer

A clean learning-focused HTTP load balancer written in Go.

This project demonstrates how to build a basic reverse-proxy load balancer with:

- round-robin request distribution
- backend health tracking
- active health checks using `/health`
- thread-safe backend selection
- graceful failure when no backends are available
- unit tests using `httptest`

This is a solid educational implementation, not a production-grade replacement for NGINX, HAProxy, Envoy, or Traefik.

---

## Features

- **Reverse proxy forwarding** using `net/http/httputil`
- **Round-robin balancing** across healthy backends
- **Atomic counters** for concurrency-safe backend selection
- **Active health checks** to detect dead backends
- **Automatic backend exclusion** when a backend becomes unhealthy
- **503 response** when no healthy backend is available
- **Unit tests** for routing and failure scenarios

---

## Project Structure

```text
loadbalancer/
├── go.mod
├── cmd/
│   ├── lb/
│   │   └── main.go
│   └── backend/
│       └── main.go
└── internal/
    └── lb/
        ├── backend.go
        ├── health.go
        ├── load_balancer.go
        └── load_balancer_test.go
````

### Structure overview

* `cmd/lb/main.go`
  Entry point for the load balancer.

* `cmd/backend/main.go`
  Small fake backend server for local testing.

* `internal/lb/backend.go`
  Backend model and reverse proxy configuration.

* `internal/lb/load_balancer.go`
  Core load balancer logic and backend selection.

* `internal/lb/health.go`
  Periodic active health checks.

* `internal/lb/load_balancer_test.go`
  Unit tests for round-robin and proxy behavior.

---

## How It Works

The load balancer keeps a list of backend servers.

For each incoming request:

1. It picks the next healthy backend using round-robin.
2. It forwards the request using a reverse proxy.
3. If the backend fails, it is marked unhealthy.
4. A background health checker periodically calls `/health` on every backend.
5. Healthy backends rejoin the rotation automatically.

If all backends are unhealthy, the load balancer returns:

```http
503 Service Unavailable
```


## Requirements

* Go 1.22 or newer


## Setup

Clone the project and move into the directory:

```bash
git clone <your-repo-url>
cd loadbalancer
```

Initialize dependencies if needed:

```bash
go mod tidy
```

## Running Locally

You need to start three fake backend servers and then the load balancer.

### 1. Start backend 1

```bash
go run ./cmd/backend -port=9001 -name=backend-1
```

### 2. Start backend 2

```bash
go run ./cmd/backend -port=9002 -name=backend-2
```

### 3. Start backend 3

```bash
go run ./cmd/backend -port=9003 -name=backend-3
```

### 4. Start the load balancer

```bash
go run ./cmd/lb
```

You should see output similar to:

```text
load balancer listening on http://localhost:8000
```

## Testing the Load Balancer

Send requests to the balancer:

```bash
curl localhost:8000
curl localhost:8000
curl localhost:8000
curl localhost:8000
```

Expected output:

```text
response from backend-1 on port 9001
response from backend-2 on port 9002
response from backend-3 on port 9003
response from backend-1 on port 9001
```

That proves round-robin is working.

## Testing Backend Failure

Kill one backend process, for example the server on port `9002`.

Now keep hitting:

```bash
curl localhost:8000
```

After the next health check cycle, the load balancer should stop forwarding traffic to that backend and continue using the healthy ones.

If you kill all backends and then run:

```bash
curl -i localhost:8000
```

You should get:

```http
HTTP/1.1 503 Service Unavailable
```

That is the correct behavior.

## Running Tests

Run all tests:

```bash
go test ./...
```

Run tests with the race detector:

```bash
go test -race ./...
```

The race detector matters because load balancers are concurrent systems. If you skip this, you are not really testing anything important.

## Health Check Behavior

Each backend exposes:

```text
/health
```

The load balancer periodically calls:

```text
http://<backend-host>/health
```

A backend is considered healthy only if it returns:

* HTTP `200 OK`

Anything else marks it unhealthy.

## Example Local Architecture

```text
Client
  |
  v
localhost:8000  (Load Balancer)
  |      |      |
  v      v      v
9001   9002   9003  (Fake Backends)
```

## Important Notes

### This is not production-ready

This project is intentionally simple and educational.

It does **not** yet include:

* graceful shutdown
* passive failure thresholds
* retries for idempotent requests
* weighted round robin
* Prometheus metrics
* request tracing
* circuit breaking
* rate limiting
* sticky sessions
* TLS termination
* dynamic backend discovery

If you need a production-grade load balancer, use battle-tested tools like:

* NGINX
* HAProxy
* Envoy
* Traefik
* Caddy

### Why build this then?

Because writing one yourself is a good way to understand:

* reverse proxies
* backend health management
* concurrent request routing
* transport tuning
* failure handling
* testable Go HTTP services

## Next Improvements

Good next steps for this project:

1. Add graceful shutdown with `server.Shutdown`
2. Add passive health failure counters
3. Add retry logic for safe requests
4. Add `/metrics` endpoint
5. Add weighted round-robin support
6. Add structured logging
7. Add config file or environment-based backend configuration

## Example Commands

Start all backends:

```bash
go run ./cmd/backend -port=9001 -name=backend-1
go run ./cmd/backend -port=9002 -name=backend-2
go run ./cmd/backend -port=9003 -name=backend-3
```

Start the balancer:

```bash
go run ./cmd/lb
```

Test:

```bash
curl localhost:8000
```

Test with race detector:

```bash
go test -race ./...
```

## License

MIT