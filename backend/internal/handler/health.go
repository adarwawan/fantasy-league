package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"fantasy-league/internal/store"
)

// SyncHealthHandler serves GET /health/sync — a freshness probe for external
// monitors (Healthchecks.io / UptimeRobot). It reports, per monitored game, how
// long ago the last fully successful sync completed, and returns 503 when any
// monitored game is stale (or has never synced) so a dumb HTTP monitor can page
// before stale data reaches a deadline.
type SyncHealthHandler struct {
	cache  cacheStore
	games  []string
	maxAge time.Duration
	now    func() time.Time
}

func NewSyncHealthHandler(c cacheStore, games []string, maxAge time.Duration) *SyncHealthHandler {
	return &SyncHealthHandler{cache: c, games: games, maxAge: maxAge, now: func() time.Time { return time.Now().UTC() }}
}

type gameSyncStatus struct {
	Game        string     `json:"game"`
	LastSuccess *time.Time `json:"last_success"`
	AgeSeconds  *float64   `json:"age_seconds"`
	Stale       bool       `json:"stale"`
}

type syncHealthResponse struct {
	OK         bool             `json:"ok"`
	MaxAgeSecs float64          `json:"max_age_seconds"`
	CheckedAt  time.Time        `json:"checked_at"`
	Games      []gameSyncStatus `json:"games"`
}

func (h *SyncHealthHandler) Sync(w http.ResponseWriter, r *http.Request) {
	now := h.now()
	resp := syncHealthResponse{
		OK:         true,
		MaxAgeSecs: h.maxAge.Seconds(),
		CheckedAt:  now,
		Games:      make([]gameSyncStatus, 0, len(h.games)),
	}

	for _, game := range h.games {
		gs := gameSyncStatus{Game: game, Stale: true}

		if last, ok := h.lastSuccess(r.Context(), game); ok {
			age := now.Sub(last)
			ageSecs := age.Seconds()
			lastCopy := last
			gs.LastSuccess = &lastCopy
			gs.AgeSeconds = &ageSecs
			gs.Stale = age > h.maxAge
		}

		if gs.Stale {
			resp.OK = false
		}
		resp.Games = append(resp.Games, gs)
	}

	status := http.StatusOK
	if !resp.OK {
		status = http.StatusServiceUnavailable
	}
	respondJSON(w, status, resp)
}

// lastSuccess reads a game's last successful sync time from cache. Returns
// ok=false when no status has been recorded yet or the read/parse fails —
// callers treat that as stale.
func (h *SyncHealthHandler) lastSuccess(ctx context.Context, game string) (time.Time, bool) {
	raw, err := h.cache.Get(ctx, store.CacheKey(game, "sync_status"))
	if err != nil || raw == nil {
		return time.Time{}, false
	}
	var payload struct {
		LastSuccess time.Time `json:"last_success"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil || payload.LastSuccess.IsZero() {
		return time.Time{}, false
	}
	return payload.LastSuccess, true
}
