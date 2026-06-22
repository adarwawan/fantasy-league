package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"fantasy-league/internal/store"
)

// DeadlineHandler handles GET /api/{game}/deadline.
// Data is populated by the syncer and read from cache — no live source calls.
type DeadlineHandler struct {
	cache cacheStore
}

func NewDeadlineHandler(c cacheStore) *DeadlineHandler {
	return &DeadlineHandler{cache: c}
}

func (h *DeadlineHandler) Deadline(w http.ResponseWriter, r *http.Request) {
	game := chi.URLParam(r, "game")
	cacheKey := store.CacheKey(game, "deadline")

	cached, err := h.cache.Get(r.Context(), cacheKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "cache read failed")
		return
	}
	if cached == nil {
		respondError(w, http.StatusServiceUnavailable, "deadline not yet available, sync pending")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "HIT")
	w.Write(cached)
}
