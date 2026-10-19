// Example 3: SERVICE MAIN — CONFIG FROM ENV, TIMEOUTS, GRACEFUL SHUTDOWN.
//
// Interview points:
//   - 12-factor: config (and secrets!) come from environment, never hard-coded.
//     In AWS the env/secret is injected by ECS/EKS from Secrets Manager, or read at startup
//     with the AWS SDK (we'll do that in the CI/CD + secrets step).
//   - Always set server timeouts (default http.Server has none -> slowloris risk).
//   - Graceful shutdown: catch SIGTERM (what Kubernetes/ECS sends), stop accepting,
//     let in-flight requests finish within a deadline.
//   - main() is thin; run() returns an error -> testable, single exit point.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/go-interview/internal/api"
	"github.com/example/go-interview/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	addr := getenv("ADDR", ":8080")

	h := api.NewHandler(store.NewMemory(), logger)
	srv := &http.Server{
		Addr:              addr,
		Handler:           h.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// ctx is cancelled when SIGINT or SIGTERM arrives.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh: // server failed to start
		return err
	case <-ctx.Done(): // shutdown signal
		logger.Info("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func getenv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
