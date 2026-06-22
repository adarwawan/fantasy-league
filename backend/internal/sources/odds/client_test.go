package odds_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"fantasy-league/internal/sources/odds"
)

// noopCache satisfies the Cache interface with no-ops.
type noopCache struct{}

func (noopCache) Get(_ context.Context, _ string) ([]byte, error) { return nil, nil }
func (noopCache) Set(_ context.Context, _ string, _ []byte, _ time.Duration) error {
	return nil
}

func TestFetchOdds_ParsesBettingJSON(t *testing.T) {
	fixture, err := os.ReadFile("testdata/betting.json")
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("x-requests-used", "1")
		w.Header().Set("x-requests-remaining", "499")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fixture)
	}))
	defer srv.Close()

	// Replace the base URL in the client by patching the HTTP server.
	// We inject the test server via a custom transport that rewrites the host.
	client := odds.NewClientWithBase(srv.URL, "test-key", time.Minute, noopCache{})

	cfg := odds.WCFOddsConfig
	matches, err := client.FetchOdds(context.Background(), cfg)
	if err != nil {
		t.Fatalf("FetchOdds: %v", err)
	}

	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}

	// First match assertions.
	m := matches[0]
	if m.ID != "abc123" {
		t.Errorf("ID: got %q, want %q", m.ID, "abc123")
	}
	if m.HomeTeam != "Portugal" {
		t.Errorf("HomeTeam: got %q, want %q", m.HomeTeam, "Portugal")
	}
	if m.AwayTeam != "DR Congo" {
		t.Errorf("AwayTeam: got %q, want %q", m.AwayTeam, "DR Congo")
	}
	if len(m.Bookmakers) != 2 {
		t.Errorf("bookmakers: got %d, want 2", len(m.Bookmakers))
	}

	bk := m.Bookmakers[0]
	if bk.Key != "bet365" {
		t.Errorf("bookmaker key: got %q, want %q", bk.Key, "bet365")
	}
	if len(bk.Markets) != 2 {
		t.Errorf("markets: got %d, want 2", len(bk.Markets))
	}

	h2h := bk.Markets[0]
	if h2h.Key != "h2h" {
		t.Errorf("market key: got %q, want %q", h2h.Key, "h2h")
	}
	if len(h2h.Outcomes) != 3 {
		t.Errorf("h2h outcomes: got %d, want 3", len(h2h.Outcomes))
	}

	totals := bk.Markets[1]
	if totals.Key != "totals" {
		t.Errorf("market key: got %q, want %q", totals.Key, "totals")
	}
	if len(totals.Outcomes) != 2 {
		t.Errorf("totals outcomes: got %d, want 2", len(totals.Outcomes))
	}
	if totals.Outcomes[0].Description != "2.5" {
		t.Errorf("totals description: got %q, want %q", totals.Outcomes[0].Description, "2.5")
	}
}

func TestFetchOdds_CacheHit(t *testing.T) {
	fixture, _ := os.ReadFile("testdata/betting.json")
	var matches []odds.OddsMatch
	_ = json.Unmarshal(fixture, &matches)

	var called int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	defer srv.Close()

	cache := &inMemCache{data: map[string][]byte{}}
	client := odds.NewClientWithBase(srv.URL, "test-key", time.Minute, cache)

	cfg := odds.WCFOddsConfig

	// First call — populates cache.
	if _, err := client.FetchOdds(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	// Second call — should hit cache, no HTTP request.
	if _, err := client.FetchOdds(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}

	if called != 1 {
		t.Errorf("expected 1 HTTP call, got %d", called)
	}
}

func TestFetchOdds_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := odds.NewClientWithBase(srv.URL, "bad-key", time.Minute, noopCache{})
	_, err := client.FetchOdds(context.Background(), odds.WCFOddsConfig)
	if err == nil {
		t.Fatal("expected error for 401, got nil")
	}
}

// inMemCache is an in-memory Cache for testing.
type inMemCache struct {
	data map[string][]byte
}

func (c *inMemCache) Get(_ context.Context, key string) ([]byte, error) {
	return c.data[key], nil
}
func (c *inMemCache) Set(_ context.Context, key string, val []byte, _ time.Duration) error {
	c.data[key] = val
	return nil
}
