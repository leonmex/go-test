package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ==========================================
// 1. DOMAIN MODELS & REPOSITORY
// ==========================================

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// UserStore acts as an in-memory thread-safe database/cache.
type UserStore struct {
	mu    sync.RWMutex
	users map[string]User
}

func NewUserStore() *UserStore {
	return &UserStore{
		users: make(map[string]User),
	}
}

func (s *UserStore) Create(user User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[user.ID]; exists {
		return errors.New("user already exists")
	}
	s.users[user.ID] = user
	return nil
}

func (s *UserStore) Get(id string) (User, bool) {
	s.mu.RLock() // Multiple readers can read concurrently
	defer s.mu.RUnlock()

	user, ok := s.users[id]
	return user, ok
}

// ==========================================
// 2. BACKGROUND WORKER POOL
// ==========================================

// Job represents background tasks (e.g., sending emails, indexing search).
type Job struct {
	UserID string
	Action string
}

type BackgroundWorker struct {
	jobs chan Job
}

func NewBackgroundWorker(bufferSize int) *BackgroundWorker {
	return &BackgroundWorker{
		jobs: make(chan Job, bufferSize),
	}
}

func (w *BackgroundWorker) Start(workerCount int) {
	for i := 1; i <= workerCount; i++ {
		go func(workerID int) {
			for job := range w.jobs {
				time.Sleep(200 * time.Millisecond)
				log.Printf("[Worker %d] Processed async task '%s' for user: %s", workerID, job.Action, job.UserID)
			}
		}(i)
	}
}

func (w *BackgroundWorker) Enqueue(job Job) {
	select {
	case w.jobs <- job:
	default:
		log.Println("[Warning] Job queue full, task dropped or requires backpressure policy")
	}
}

// ==========================================
// 3. HTTP HANDLERS & MIDDLEWARE
// ==========================================

type Server struct {
	store  *UserStore
	worker *BackgroundWorker
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	if payload.ID == "" || payload.Email == "" {
		http.Error(w, `{"error":"id and email are required"}`, http.StatusUnprocessableEntity)
		return
	}

	user := User{
		ID:        payload.ID,
		Email:     payload.Email,
		Name:      payload.Name,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.store.Create(user); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusConflict)
		return
	}

	// Offload slow operations to worker pool
	s.worker.Enqueue(Job{UserID: user.ID, Action: "SendWelcomeEmail"})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	user, found := s.store.Get(id)
	if !found {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// ==========================================
// 4. MAIN & GRACEFUL SHUTDOWN
// ==========================================

func main() {
	store := NewUserStore()
	worker := NewBackgroundWorker(100)
	worker.Start(3)

	srv := &Server{
		store:  store,
		worker: worker,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("POST /api/v1/users", srv.handleCreateUser)
	mux.HandleFunc("GET /api/v1/users/{id}", srv.handleGetUser)

	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		log.Println("Server running on http://localhost:8080")
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Forced shutdown error: %v", err)
	}
	log.Println("Server stopped cleanly.")
}
