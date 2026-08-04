package trends

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// Cache is the subset of store.Cache used for leader-board response caching.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
}

// leaderCacheTTL keeps the "fastest movers" board cheap under the near-deadline
// read spike: everyone sees the same board, so one DB query per TTL serves all.
const leaderCacheTTL = 45 * time.Second

const defaultLeaderLimit = 25

// Handler serves the read API under /api/trends plus the secret-guarded arm/poll
// controls. Poller is optional (nil-safe) so tests can construct without one.
type Handler struct {
	store  *Store
	cache  Cache
	client *Client
	poller *Poller
	secret string
}

func NewHandler(store *Store, cache Cache, client *Client, poller *Poller, secret string) *Handler {
	return &Handler{store: store, cache: cache, client: client, poller: poller, secret: secret}
}

// Session — GET /api/trends/session. Returns the active window or {active:false}.
func (h *Handler) Session(w http.ResponseWriter, r *http.Request) {
	sess, err := h.store.ActiveSession(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	if sess == nil {
		respondJSON(w, http.StatusOK, map[string]any{"active": false})
		return
	}
	respondJSON(w, http.StatusOK, sess)
}

// Leaders — GET /api/trends/leaders?window=30m&limit=25. Needs an active session.
func (h *Handler) Leaders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sess, err := h.store.ActiveSession(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	if sess == nil {
		respondJSON(w, http.StatusOK, map[string]any{"active": false, "leaders": []LeaderRow{}})
		return
	}

	window := parseWindow(r.URL.Query().Get("window"), 30*time.Minute)
	limit := parseLimit(r.URL.Query().Get("limit"), defaultLeaderLimit)
	dir := ParseDirection(r.URL.Query().Get("direction"))
	metric := MetricForGameweek(sess.Gameweek)

	cacheKey := "trends:leaders:" + strconv.Itoa(sess.Gameweek) + ":" +
		string(metric) + ":" + string(dir) + ":" + window.String() + ":" + strconv.Itoa(limit)
	if b := h.cacheGet(ctx, cacheKey); b != nil {
		writeRaw(w, b)
		return
	}

	leaders, err := h.store.Leaders(ctx, sess.Gameweek, limit, window, dir, metric)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	if leaders == nil {
		leaders = []LeaderRow{}
	}
	resp := map[string]any{
		"active":    true,
		"gameweek":  sess.Gameweek,
		"window":    window.String(),
		"direction": string(dir),
		"metric":    string(metric),
		"leaders":   leaders,
	}
	b, _ := json.Marshal(resp)
	h.cacheSet(ctx, cacheKey, b)
	writeRaw(w, b)
}

// Series — GET /api/trends/player/{extId}/series.
func (h *Handler) Series(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	extID, err := strconv.Atoi(chi.URLParam(r, "extId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid player id")
		return
	}
	sess, err := h.store.ActiveSession(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	if sess == nil {
		respondJSON(w, http.StatusOK, map[string]any{"active": false, "series": []SeriesPoint{}})
		return
	}
	series, err := h.store.Series(ctx, sess.Gameweek, extID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	if series == nil {
		series = []SeriesPoint{}
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"active":        true,
		"gameweek":      sess.Gameweek,
		"player_ext_id": extID,
		"series":        series,
	})
}

// Arm — POST /api/trends/session (secret-guarded). Body {gameweek, deadline} is
// optional; when absent, the next FPL deadline is auto-detected.
func (h *Handler) Arm(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	ctx := r.Context()

	var body struct {
		Gameweek int    `json:"gameweek"`
		Deadline string `json:"deadline"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	gw, deadline := body.Gameweek, time.Time{}
	if body.Deadline != "" {
		if t, err := time.Parse(time.RFC3339, body.Deadline); err == nil {
			deadline = t
		}
	}
	if gw == 0 || deadline.IsZero() {
		autoGW, autoDL, err := h.client.NextDeadline(ctx)
		if err != nil {
			respondError(w, http.StatusBadGateway, "could not auto-detect deadline")
			return
		}
		if gw == 0 {
			gw = autoGW
		}
		if deadline.IsZero() {
			deadline = autoDL
		}
	}

	if err := h.store.ArmSession(ctx, gw, deadline); err != nil {
		respondError(w, http.StatusInternalServerError, "arm failed")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"armed": true, "gameweek": gw, "deadline": deadline})
}

// Poll — POST /api/trends/poll (secret-guarded). Triggers one capture now, so a
// session can be populated on demand (local dev, or a manual top-up).
func (h *Handler) Poll(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if h.poller == nil {
		respondError(w, http.StatusServiceUnavailable, "poller not configured")
		return
	}
	h.poller.Tick(r.Context())
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("polled"))
}

func (h *Handler) authorized(r *http.Request) bool {
	return h.secret != "" && r.Header.Get("X-Sync-Secret") == h.secret
}

func parseWindow(s string, def time.Duration) time.Duration {
	if s == "" {
		return def
	}
	if d, err := time.ParseDuration(s); err == nil && d > 0 {
		return d
	}
	return def
}

func parseLimit(s string, def int) int {
	if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 200 {
		return n
	}
	return def
}

func (h *Handler) cacheGet(ctx context.Context, key string) []byte {
	if h.cache == nil {
		return nil
	}
	b, err := h.cache.Get(ctx, key)
	if err != nil {
		return nil
	}
	return b
}

func (h *Handler) cacheSet(ctx context.Context, key string, b []byte) {
	if h.cache != nil {
		_ = h.cache.Set(ctx, key, b, leaderCacheTTL)
	}
}

func writeRaw(w http.ResponseWriter, b []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}
