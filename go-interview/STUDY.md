# Go Services — Interview Study Guide (Step 1 of 4)

Roadmap: **1. Go examples (this)** → 2. Tests (unit, table-driven, mocks, -race) →
3. GitHub Actions CI/CD + AWS secrets (OIDC, Secrets Manager) → 4. Blackbox testing.

## Project layout (say this in the interview)
```
cmd/server/main.go        entrypoint: config, wiring, graceful shutdown
internal/api/handler.go   HTTP layer (handlers, routing, JSON)
internal/store/store.go   data layer (interface + implementation)
examples/workerpool/      concurrency pattern
```
`internal/` = cannot be imported by other modules. `cmd/` = one folder per binary.

## Example 1 — Interfaces + Mutex (store.go)
Memorize the shape:
```go
type Store interface {
    Get(ctx context.Context, id string) (Item, error)
    Put(ctx context.Context, item Item) error
}
type Memory struct { mu sync.RWMutex; items map[string]Item }
```
- "Accept interfaces, return structs." Interfaces are satisfied **implicitly**.
- Maps are not concurrency-safe → `RWMutex` (RLock for reads, Lock for writes).
- Sentinel error `ErrNotFound` + `errors.Is` in callers.

## Example 2 — HTTP handlers (handler.go)
```go
mux.HandleFunc("GET /items/{id}", h.getItem)   // Go 1.22+
id := r.PathValue("id")
```
- Dependencies injected through `NewHandler(store, logger)`.
- `http.MaxBytesReader` + `DisallowUnknownFields` = input hardening.
- Map errors to status codes: not found → 404, bad input → 400, else 500 (never leak internals).

## Example 3 — main.go: config, timeouts, graceful shutdown
```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
go srv.ListenAndServe()
<-ctx.Done()
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
srv.Shutdown(shutdownCtx)
```
- Config & secrets from **environment** (12-factor), never in code or git.
- Always set `ReadHeaderTimeout`/`ReadTimeout`/`WriteTimeout` (defaults are infinite).
- Kubernetes/ECS send **SIGTERM** → finish in-flight requests, then exit.

## Example 4 — Worker pool (workerpool/main.go)
Three goroutine roles: **producer** (fills `jobs`, closes it), **workers** (range jobs),
**closer** (`wg.Wait()` then `close(results)`).
- The sender closes the channel. Sending on a closed channel panics.
- `select { case <-ctx.Done(): return; case ch <- v: }` prevents goroutine leaks.
- Run with `go test -race` / `go run -race`.

## Rapid-fire questions
| Question | Short answer |
|---|---|
| Goroutine vs thread? | Goroutine: ~2 KB stack, scheduled by Go runtime (M:N), cheap to create. |
| Buffered vs unbuffered channel? | Unbuffered: send blocks until receive. Buffered: blocks only when full. |
| What is `context` for? | Cancellation, deadlines, request-scoped values across API boundaries. |
| `defer` order? | LIFO; args evaluated at defer time. |
| Value vs pointer receiver? | Pointer when method mutates or struct is large / has a mutex. |
| How to avoid data races? | Mutex, channels, `sync/atomic`; detect with `-race`. |
| Error wrapping? | `fmt.Errorf("get user: %w", err)` then `errors.Is` / `errors.As`. |
| nil interface trap? | Interface holding a nil pointer is **not** nil (type is set). |
| Slice vs array? | Array fixed size value; slice = pointer+len+cap view, `append` may reallocate. |
| How do you handle secrets? | Env/Secrets Manager at runtime, IAM role (no static keys), never log them. |

## Commands
```
go run ./cmd/server           # start service on :8080
go run ./examples/workerpool
go vet ./... && go build ./...
curl localhost:8080/health
curl -XPOST localhost:8080/items -d '{"id":"1","name":"book"}'
curl localhost:8080/items/1
```
