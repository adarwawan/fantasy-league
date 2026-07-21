package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"fantasy-league/internal/store"
)

// mapCache is a minimal in-memory cacheStore for handler tests.
type mapCache struct{ m map[string][]byte }

func newMapCache() *mapCache { return &mapCache{m: map[string][]byte{}} }

func (c *mapCache) Get(_ context.Context, key string) ([]byte, error) { return c.m[key], nil }
func (c *mapCache) Set(_ context.Context, key string, val []byte, _ time.Duration) error {
	c.m[key] = val
	return nil
}

func (c *mapCache) setSyncStatus(game string, last time.Time) {
	b, _ := json.Marshal(map[string]any{"game": game, "last_success": last})
	c.m[store.CacheKey(game, "sync_status")] = b
}

func decodeHealth(t *testing.T, body []byte) syncHealthResponse {
	t.Helper()
	var resp syncHealthResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp
}

func TestSyncHealth_FreshIsOK(t *testing.T) {
	cache := newMapCache()
	now := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
	cache.setSyncStatus("fpl", now.Add(-30*time.Minute))

	h := NewSyncHealthHandler(cache, []string{"fpl"}, 26*time.Hour)
	h.now = func() time.Time { return now }

	rr := httptest.NewRecorder()
	h.Sync(rr, httptest.NewRequest(http.MethodGet, "/health/sync", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	resp := decodeHealth(t, rr.Body.Bytes())
	if !resp.OK || len(resp.Games) != 1 || resp.Games[0].Stale {
		t.Fatalf("expected healthy fresh game, got %+v", resp)
	}
	if resp.Games[0].AgeSeconds == nil || *resp.Games[0].AgeSeconds != 1800 {
		t.Fatalf("age = %v, want 1800s", resp.Games[0].AgeSeconds)
	}
}

func TestSyncHealth_StaleReturns503(t *testing.T) {
	cache := newMapCache()
	now := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
	cache.setSyncStatus("fpl", now.Add(-30*time.Hour)) // older than 26h

	h := NewSyncHealthHandler(cache, []string{"fpl"}, 26*time.Hour)
	h.now = func() time.Time { return now }

	rr := httptest.NewRecorder()
	h.Sync(rr, httptest.NewRequest(http.MethodGet, "/health/sync", nil))

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rr.Code)
	}
	resp := decodeHealth(t, rr.Body.Bytes())
	if resp.OK || !resp.Games[0].Stale {
		t.Fatalf("expected stale/unhealthy, got %+v", resp)
	}
}

func TestSyncHealth_NeverSyncedIsStale(t *testing.T) {
	cache := newMapCache() // no status recorded
	now := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)

	h := NewSyncHealthHandler(cache, []string{"fpl"}, 26*time.Hour)
	h.now = func() time.Time { return now }

	rr := httptest.NewRecorder()
	h.Sync(rr, httptest.NewRequest(http.MethodGet, "/health/sync", nil))

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rr.Code)
	}
	resp := decodeHealth(t, rr.Body.Bytes())
	if resp.Games[0].LastSuccess != nil || !resp.Games[0].Stale {
		t.Fatalf("expected never-synced game marked stale with nil last_success, got %+v", resp.Games[0])
	}
}

func TestSyncHealth_OneStaleGameFailsOverall(t *testing.T) {
	cache := newMapCache()
	now := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
	cache.setSyncStatus("fpl", now.Add(-1*time.Hour))  // fresh
	cache.setSyncStatus("wcf", now.Add(-40*time.Hour)) // stale

	h := NewSyncHealthHandler(cache, []string{"fpl", "wcf"}, 26*time.Hour)
	h.now = func() time.Time { return now }

	rr := httptest.NewRecorder()
	h.Sync(rr, httptest.NewRequest(http.MethodGet, "/health/sync", nil))

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503 when any game stale", rr.Code)
	}
}
