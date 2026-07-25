package setpiece

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const understatBase = "https://understat.com"

// Understat's AJAX endpoints require this header and return gzip'd JSON.
const xRequestedWith = "XMLHttpRequest"

// Cache is the subset of store.Cache used by the client (raw-payload caching).
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
}

// Client fetches Understat league and match data. It is deliberately small and
// self-contained so the scraping dependency degrades independently of the main
// app (see docs §3, §7).
type Client struct {
	baseURL  string
	cacheTTL time.Duration
	cache    Cache
	http     *http.Client
}

func NewClient(cacheTTL time.Duration, cache Cache) *Client {
	return NewClientWithBase(understatBase, cacheTTL, cache)
}

// NewClientWithBase overrides the base URL (used in tests).
func NewClientWithBase(baseURL string, cacheTTL time.Duration, cache Cache) *Client {
	return &Client{
		baseURL:  baseURL,
		cacheTTL: cacheTTL,
		cache:    cache,
		http:     &http.Client{Timeout: 20 * time.Second},
	}
}

// FinishedMatches returns the ids of finished EPL matches for the given season
// (season = starting year, e.g. "2025" for 2025/26).
func (c *Client) FinishedMatches(ctx context.Context, season string) ([]string, error) {
	url := fmt.Sprintf("%s/getLeagueData/EPL/%s", c.baseURL, season)
	body, err := c.get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("getLeagueData EPL/%s: %w", season, err)
	}
	// getLeagueData returns {teams, players, dates}; the fixture list is `dates`.
	var payload leagueData
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode league data: %w", err)
	}
	ids := make([]string, 0, len(payload.Dates))
	for _, m := range payload.Dates {
		if m.IsResult {
			ids = append(ids, m.ID)
		}
	}
	return ids, nil
}

// MatchShots returns every shot (home + away) for a match. The raw payload is
// cached in Redis under "setpiece:match:{id}" — finished matches are immutable,
// so a long TTL avoids re-fetching on every daily sync.
func (c *Client) MatchShots(ctx context.Context, matchID string) ([]shot, error) {
	cacheKey := "setpiece:match:" + matchID
	if c.cache != nil {
		if b, err := c.cache.Get(ctx, cacheKey); err == nil && b != nil {
			if shots, err := decodeShots(b); err == nil {
				return shots, nil
			}
		}
	}

	url := fmt.Sprintf("%s/getMatchData/%s", c.baseURL, matchID)
	body, err := c.get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("getMatchData %s: %w", matchID, err)
	}

	shots, err := decodeShots(body)
	if err != nil {
		return nil, fmt.Errorf("decode match %s: %w", matchID, err)
	}

	if c.cache != nil {
		_ = c.cache.Set(ctx, cacheKey, body, c.cacheTTL)
	}
	return shots, nil
}

func decodeShots(body []byte) ([]shot, error) {
	var md matchData
	if err := json.Unmarshal(body, &md); err != nil {
		return nil, err
	}
	shots := make([]shot, 0, len(md.Shots.H)+len(md.Shots.A))
	shots = append(shots, md.Shots.H...)
	shots = append(shots, md.Shots.A...)
	return shots, nil
}

// get performs a GET with the Understat AJAX headers and transparently
// gunzips the response when needed.
func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Requested-With", xRequestedWith)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; fantasy-league/1.0)")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("gzip: %w", err)
		}
		defer gz.Close()
		reader = gz
	}
	return io.ReadAll(reader)
}
