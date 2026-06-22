package odds

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

const oddsAPIBase = "https://api.the-odds-api.com/v4"

// Cache is the subset of store.Cache used by the client.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
}

type Client struct {
	baseURL  string
	apiKey   string
	cacheTTL time.Duration
	cache    Cache
	http     *http.Client
}

func NewClient(apiKey string, cacheTTL time.Duration, cache Cache) *Client {
	return NewClientWithBase(oddsAPIBase, apiKey, cacheTTL, cache)
}

// NewClientWithBase allows overriding the base URL (used in tests).
func NewClientWithBase(baseURL, apiKey string, cacheTTL time.Duration, cache Cache) *Client {
	return &Client{
		baseURL:  baseURL,
		apiKey:   apiKey,
		cacheTTL: cacheTTL,
		cache:    cache,
		http:     &http.Client{Timeout: 15 * time.Second},
	}
}

// FetchOdds returns raw OddsMatch records for the given game config.
// Results are cached in Redis under "{gameID}:odds:raw".
func (c *Client) FetchOdds(ctx context.Context, cfg GameOddsConfig) ([]OddsMatch, error) {
	cacheKey := cfg.GameID + ":odds:raw"

	if c.cache != nil {
		if b, err := c.cache.Get(ctx, cacheKey); err == nil && b != nil {
			var matches []OddsMatch
			if err := json.Unmarshal(b, &matches); err == nil {
				return matches, nil
			}
		}
	}

	url := fmt.Sprintf(
		"%s/sports/%s/odds?regions=eu&markets=h2h,totals&oddsFormat=decimal&apiKey=%s",
		c.baseURL, cfg.SportKey, c.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET odds (%s): %w", cfg.SportKey, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET odds (%s): status %d", cfg.SportKey, resp.StatusCode)
	}

	logQuota(resp)

	var matches []OddsMatch
	if err := json.NewDecoder(resp.Body).Decode(&matches); err != nil {
		return nil, fmt.Errorf("decode odds (%s): %w", cfg.SportKey, err)
	}

	if c.cache != nil {
		if b, err := json.Marshal(matches); err == nil {
			_ = c.cache.Set(ctx, cacheKey, b, c.cacheTTL)
		}
	}

	return matches, nil
}

func logQuota(resp *http.Response) {
	remaining := resp.Header.Get("x-requests-remaining")
	used := resp.Header.Get("x-requests-used")
	if remaining == "" && used == "" {
		return
	}
	rem, _ := strconv.Atoi(remaining)
	u, _ := strconv.Atoi(used)
	log.Printf("odds api quota: %d used, %d remaining", u, rem)
}
