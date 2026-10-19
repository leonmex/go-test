# Golang Backend Architecture Starter

Starter backend implementation in Go demonstrating key concepts from the backend architecture fundamentals video:

- **RESTful Routing**: Using Go 1.22+ `net/http` route matching (`POST /api/v1/users`, `GET /api/v1/users/{id}`).
- **Thread-Safe In-Memory Store**: `sync.RWMutex` protecting concurrent read/write access.
- **Worker Pool / Async Processing**: Goroutines and buffered channels handling background jobs without blocking request handlers.
- **Middleware**: Request logging and timing.
- **Graceful Shutdown**: Handling system signals (`SIGINT`, `SIGTERM`) with context timeouts to cleanly drain active requests.

## How to Run

```bash
# Run server
go run main.go

# Test creating a user
curl -i -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"id":"u1","email":"alice@example.com","name":"Alice"}'

# Test fetching a user
curl -i http://localhost:8080/api/v1/users/u1
```
