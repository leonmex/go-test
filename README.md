# Go Projects Docker Setup

This workspace contains two distinct Go projects managed through a single Docker Compose setup using the latest stable Go runtime (Go 1.24) and Air for live hot-reloading:

1. `go-api-images`: Backend API with user management, in-memory store, and background worker queue.
2. `go-interview`: Production-ready Go service demonstrating clean architecture, graceful shutdown, store interfaces, and interview patterns.

## Port Bindings

The active host ports were verified to prevent conflicts with any running containers on the host machine:

| Service | Container Port | Host Port | Environment Variable | Health Check Endpoint | Default URL |
|---|---|---|---|---|---|
| `go-api-images` | `8080` | `8080` | `GO_API_IMAGES_PORT` | `http://localhost:8080/health` | `http://localhost:8080` |
| `go-interview` | `8080` | `8081` | `GO_INTERVIEW_PORT` | `http://localhost:8080/health` | `http://localhost:8081` |

Host ports can be customized at any time in [.env](file:///home/noel/projects/docker/Go-TS-Python/gointerview/.env).

## Container Health Checks

Both services feature automated Docker health checks:
- **Test Command**: `curl -f http://localhost:8080/health || exit 1`
- **Interval**: 10 seconds
- **Timeout**: 3 seconds
- **Retries**: 3
- **Start Period**: 5 seconds (grace period for initial Air build)

Check health status with:
```bash
docker compose ps
```

## IDE Read and Write Sync

- Both project directories are bind-mounted (`./go-api-images:/app` and `./go-interview:/app`).
- File edits made in your IDE are instantly picked up by Air and trigger an automatic rebuild.
- Build artifacts are directed to `/tmp/air/...` inside the container, preventing root-owned binaries or locks from dirtying your host workspace.
- Go module downloads and compiler caches are persisted in named Docker volumes (`go-interview-mod-cache` and `go-interview-build-cache`) for fast startup and rebuild times.

## How to Run

### Start Both Projects Together
```bash
docker compose up -d
```

### Start Only One Project
To run `go-api-images` individually:
```bash
docker compose up go-api-images
```

To run `go-interview` individually:
```bash
docker compose up go-interview
```

### View Logs
```bash
# Follow logs for all services
docker compose logs -f

# Follow logs for a specific service
docker compose logs -f go-api-images
docker compose logs -f go-interview
```

### Stop Services
```bash
docker compose down
```

## Running Tests and Shell Commands

You can run any Go command directly inside the container without installing Go on your host:

```bash
# Run unit tests in go-interview
docker compose exec go-interview go test -v ./...

# Run race detector
docker compose exec go-interview go test -race ./...

# Run the worker pool example
docker compose exec go-interview go run ./examples/workerpool/main.go

# Open an interactive shell inside a container
docker compose exec go-interview bash
docker compose exec go-api-images bash
```

## Verification Requests

### Test go-api-images (Port 8080)
```bash
# Create a user
curl -i -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"id":"u1","email":"alice@example.com","name":"Alice"}'

# Retrieve user
curl -i http://localhost:8080/api/v1/users/u1
```

### Test go-interview (Port 8081)
```bash
# Query health or items
curl -i http://localhost:8081/healthz
```
