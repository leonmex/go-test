// Package api is the HTTP layer. Example 2: HTTP HANDLERS + DEPENDENCY INJECTION.
//
// Interview points:
//   - Handler struct receives its dependencies (Store) via constructor = DI without frameworks.
//   - Go 1.22+ ServeMux supports methods and path params: "GET /items/{id}", r.PathValue("id").
//   - Validate input, map domain errors to HTTP codes (errors.Is), always set Content-Type.
//   - Limit request body size (http.MaxBytesReader) -> basic security.
package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/example/go-interview/internal/store"
)

type Handler struct {
	store  store.Store
	logger *slog.Logger
}

func NewHandler(s store.Store, l *slog.Logger) *Handler {
	return &Handler{store: s, logger: l}
}

// Routes returns the router. Tests call this too (httptest), so tests hit real routing.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /items/{id}", h.getItem)
	mux.HandleFunc("POST /items", h.createItem)
	return mux
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) getItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	it, err := h.store.Get(r.Context(), id)
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	case err != nil:
		h.logger.Error("get item", "id", id, "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal"})
		return
	}
	writeJSON(w, http.StatusOK, it)
}

func (h *Handler) createItem(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB max
	var in store.Item
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if in.ID == "" || in.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id and name required"})
		return
	}
	if err := h.store.Put(r.Context(), in); err != nil {
		h.logger.Error("put item", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal"})
		return
	}
	writeJSON(w, http.StatusCreated, in)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
