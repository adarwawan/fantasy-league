package fotmob

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	fotmobBase    = "https://www.fotmob.com"
	premierLeague = 47
	matchCacheKey = "fotmob:match:"
)

// Cache is the subset of store.Cache used by the client.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
}

// Client fetches FotMob's unofficial JSON API. It is unauthenticated today but
// undocumented, so treat shape changes as expected and fail per match, not per run.
type Client struct {
	baseURL  string
	cacheTTL time.Duration
	cache    Cache
	http     *http.Client
}

func NewClient(cacheTTL time.Duration, cache Cache) *Client {
	return NewClientWithBase(fotmobBase, cacheTTL, cache)
}

func NewClientWithBase(baseURL string, cacheTTL time.Duration, cache Cache) *Client {
	return &Client{baseURL: baseURL, cacheTTL: cacheTTL, cache: cache, http: &http.Client{Timeout: 30 * time.Second}}
}

// FinishedMatches returns ids of finished Premier League matches for a season
// given as its starting year ("2026" = 2026/2027).
func (c *Client) FinishedMatches(ctx context.Context, season string) ([]string, error) {
	y, err := time.Parse("2006", season)
	if err != nil {
		return nil, fmt.Errorf("season %q: want a starting year", season)
	}
	q := url.Values{}
	q.Set("id", fmt.Sprint(premierLeague))
	q.Set("ccode3", "GBR")
	q.Set("season", fmt.Sprintf("%d/%d", y.Year(), y.Year()+1))
	body, err := c.get(ctx, "/api/data/leagues?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("leagues: %w", err)
	}
	var p leaguePayload
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, fmt.Errorf("decode leagues: %w", err)
	}
	var ids []string
	for _, m := range p.Fixtures.AllMatches {
		// `finished` only, never the provisional state (docs/decisions/002).
		if m.Status.Finished && !m.Status.Cancelled {
			ids = append(ids, m.ID)
		}
	}
	return ids, nil
}

// Match returns a finished match's shots. Finished matches are immutable, so the
// slim form is cached for cacheTTL.
func (c *Client) Match(ctx context.Context, id string) (matchData, error) {
	if c.cache != nil {
		if b, err := c.cache.Get(ctx, matchCacheKey+id); err == nil && b != nil {
			var m matchData
			if json.Unmarshal(b, &m) == nil {
				return m, nil
			}
		}
	}
	body, err := c.get(ctx, "/api/data/matchDetails?matchId="+url.QueryEscape(id))
	if err != nil {
		return matchData{}, fmt.Errorf("matchDetails %s: %w", id, err)
	}
	m, err := decodeMatch(body)
	if err != nil {
		return matchData{}, fmt.Errorf("match %s: %w", id, err)
	}
	if c.cache != nil {
		if b, err := json.Marshal(m); err == nil {
			_ = c.cache.Set(ctx, matchCacheKey+id, b, c.cacheTTL)
		}
	}
	return m, nil
}

func decodeMatch(body []byte) (matchData, error) {
	var p matchPayload
	if err := json.Unmarshal(body, &p); err != nil {
		return matchData{}, fmt.Errorf("decode: %w", err)
	}
	if p.Content.Shotmap == nil {
		// Also what FotMob's {"error":true} body decodes to. Erroring keeps the
		// match unprocessed so a later run retries once shot data is published.
		return matchData{}, fmt.Errorf("no shotmap")
	}
	t, err := time.Parse(time.RFC3339, p.General.MatchTimeUTC)
	if err != nil {
		return matchData{}, fmt.Errorf("match date %q: %w", p.General.MatchTimeUTC, err)
	}
	return matchData{Date: t, Home: p.General.HomeTeam, Away: p.General.AwayTeam, Shots: p.Content.Shotmap.Shots}, nil
}

func (c *Client) get(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; fantasy-league/1.0)")
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	var r io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		r = gz
	}
	return io.ReadAll(r)
}
