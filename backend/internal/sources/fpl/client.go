package fpl

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const baseURL = "https://fantasy.premierleague.com/api"

// userAgent mimics a mainstream desktop browser. FPL sits behind Cloudflare,
// which throttles (and sometimes blocks) requests carrying obviously scripted
// User-Agent strings more aggressively than browser-like ones.
const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"

// maxRetries is how many times get retries a request that comes back 429 or 5xx
// before giving up. Combined with jittered backoff this lets a burst that trips
// the FPL rate limiter recover instead of failing wholesale.
const maxRetries = 4

// bootstrapTTL is how long a fetched /bootstrap-static/ response is reused
// before a fresh download. The endpoint is large (~600KB) and heavily rate
// limited; on-demand reads (e.g. CurrentGW on every planner team-load) only
// need the current gameweek, which barely changes within this window.
const bootstrapTTL = 10 * time.Minute

type client struct {
	http    *http.Client
	baseURL string

	// bootstrapCache collapses concurrent misses onto a single fetch and serves
	// a cached response for bootstrapTTL. See fetchBootstrap.
	bootMu     sync.Mutex
	bootData   *bootstrapResponse
	bootExpiry time.Time
}

func newClient() *client {
	return &client{
		http:    &http.Client{Timeout: 15 * time.Second},
		baseURL: baseURL,
	}
}

func (c *client) get(ctx context.Context, path string, dst any) error {
	var (
		lastStatus int
		retryAfter time.Duration
	)
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retrying: honor Retry-After if the server sent one,
			// otherwise use exponential backoff with full jitter so a burst of
			// workers doesn't retry in a synchronized wall.
			if err := sleep(ctx, backoff(attempt, retryAfter)); err != nil {
				return err
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Accept", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			return fmt.Errorf("GET %s: %w", path, err)
		}

		if resp.StatusCode == http.StatusOK {
			err := json.NewDecoder(resp.Body).Decode(dst)
			resp.Body.Close()
			if err != nil {
				return fmt.Errorf("decode %s: %w", path, err)
			}
			return nil
		}

		// Retry on 429 (rate limited) and transient 5xx; anything else is fatal.
		retriable := resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
		lastStatus = resp.StatusCode
		retryAfter = parseRetryAfter(resp.Header.Get("Retry-After"))
		resp.Body.Close()
		if !retriable || attempt == maxRetries {
			return fmt.Errorf("GET %s: status %d", path, resp.StatusCode)
		}
	}
	return fmt.Errorf("GET %s: exhausted retries (last status %d)", path, lastStatus)
}

// backoff returns how long to wait before the given retry attempt (1-based).
// It prefers a server-provided Retry-After, otherwise uses exponential backoff
// (0.5s, 1s, 2s, 4s...) with full jitter to desynchronize concurrent workers.
func backoff(attempt int, retryAfter time.Duration) time.Duration {
	if retryAfter > 0 {
		// Add a little jitter on top so many workers don't wake simultaneously.
		return retryAfter + time.Duration(rand.Int63n(int64(500*time.Millisecond)))
	}
	base := (500 * time.Millisecond) << (attempt - 1)
	return time.Duration(rand.Int63n(int64(base))) + base/2
}

// parseRetryAfter reads a Retry-After header expressed in delta-seconds. HTTP
// dates are ignored (FPL sends seconds); an unparseable value yields 0.
func parseRetryAfter(v string) time.Duration {
	if v == "" {
		return 0
	}
	secs, err := strconv.Atoi(v)
	if err != nil || secs < 0 {
		return 0
	}
	return time.Duration(secs) * time.Second
}

// sleep waits for d or until ctx is cancelled, whichever comes first.
func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// --- API response types ---

type bootstrapResponse struct {
	Elements []fplPlayer `json:"elements"`
	Teams    []fplTeam   `json:"teams"`
	Events   []fplEvent  `json:"events"`
}

type fplPlayer struct {
	ID                int     `json:"id"`
	WebName           string  `json:"web_name"`
	Team              int     `json:"team"`
	ElementType       int     `json:"element_type"`
	NowCost           int     `json:"now_cost"`
	Form              string  `json:"form"`
	SelectedByPercent string  `json:"selected_by_percent"`
	Status            string  `json:"status"`
	News              string  `json:"news"`
	Minutes           int     `json:"minutes"`
	// Set-piece taker ranks; the API sends null for players without the duty.
	PenaltiesOrder       *int `json:"penalties_order"`
	DirectFreekicksOrder *int `json:"direct_freekicks_order"`
	CornersIndirectOrder *int `json:"corners_and_indirect_freekicks_order"`
}

type fplTeam struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
}

type fplEvent struct {
	ID           int    `json:"id"`
	IsCurrent    bool   `json:"is_current"`
	IsNext       bool   `json:"is_next"`
	DeadlineTime string `json:"deadline_time"`
}

type fplFixture struct {
	ID              int    `json:"id"`
	Event           *int   `json:"event"`
	TeamH           int    `json:"team_h"`
	TeamA           int    `json:"team_a"`
	TeamHDifficulty int    `json:"team_h_difficulty"`
	TeamADifficulty int    `json:"team_a_difficulty"`
	KickoffTime     string `json:"kickoff_time"`
	Finished        bool   `json:"finished"`
	TeamHScore      *int   `json:"team_h_score"`
	TeamAScore      *int   `json:"team_a_score"`
}

type leagueStandingsResponse struct {
	Standings struct {
		Results []fplStandingEntry `json:"results"`
		HasNext bool               `json:"has_next"`
		Page    int                `json:"page"`
	} `json:"standings"`
}

type fplStandingEntry struct {
	Entry     int    `json:"entry"`
	EntryName string `json:"entry_name"`
	Rank      int    `json:"rank"`
}

type eventLiveResponse struct {
	Elements []fplLiveElement `json:"elements"`
}

type fplLiveElement struct {
	ID    int `json:"id"`
	Stats struct {
		Minutes     int `json:"minutes"`
		Starts      int `json:"starts"`
		TotalPoints int `json:"total_points"`
		GoalsScored int `json:"goals_scored"`
		Assists     int `json:"assists"`
		Bonus       int `json:"bonus"`
		CleanSheets int `json:"clean_sheets"`
		DefCon      int `json:"defensive_contribution"`
		// ICT components arrive as decimal strings, e.g. "38.2".
		Influence  string `json:"influence"`
		Creativity string `json:"creativity"`
		Threat     string `json:"threat"`
	} `json:"stats"`
	// Explain holds the per-fixture points breakdown. During a double gameweek a
	// player has one entry per fixture, so threshold-based components (e.g.
	// defensive contribution) must be summed per fixture rather than derived from
	// the aggregate stats value, which would apply the threshold only once.
	Explain []fplExplain `json:"explain"`
}

type fplExplain struct {
	Stats []fplExplainStat `json:"stats"`
}

type fplExplainStat struct {
	Identifier string `json:"identifier"`
	Points     int    `json:"points"`
	Value      int    `json:"value"`
}

type picksResponse struct {
	Picks []fplPick `json:"picks"`
}

type fplPick struct {
	Element       int  `json:"element"`
	Position      int  `json:"position"`
	Multiplier    int  `json:"multiplier"`
	IsCaptain     bool `json:"is_captain"`
	IsViceCaptain bool `json:"is_vice_captain"`
}

// fetchBootstrap returns /bootstrap-static/, cached for bootstrapTTL. The mutex
// is held across the network fetch so concurrent misses collapse onto one
// request instead of each hammering the rate-limited endpoint. If a refresh
// fails but a previously cached response exists, that stale response is served
// rather than propagating the error.
func (c *client) fetchBootstrap(ctx context.Context) (*bootstrapResponse, error) {
	c.bootMu.Lock()
	defer c.bootMu.Unlock()

	if c.bootData != nil && time.Now().Before(c.bootExpiry) {
		return c.bootData, nil
	}

	var r bootstrapResponse
	if err := c.get(ctx, "/bootstrap-static/", &r); err != nil {
		if c.bootData != nil {
			// Serve the last good response rather than failing the caller.
			return c.bootData, nil
		}
		return nil, err
	}

	c.bootData = &r
	c.bootExpiry = time.Now().Add(bootstrapTTL)
	return c.bootData, nil
}

func (c *client) fetchFixtures(ctx context.Context) ([]fplFixture, error) {
	var r []fplFixture
	return r, c.get(ctx, "/fixtures/", &r)
}

func (c *client) fetchLeagueStandings(ctx context.Context, leagueID, maxManagers int) ([]fplStandingEntry, error) {
	var all []fplStandingEntry
	page := 1
	for len(all) < maxManagers {
		var r leagueStandingsResponse
		path := fmt.Sprintf("/leagues-classic/%d/standings/?page_standings=%d", leagueID, page)
		if err := c.get(ctx, path, &r); err != nil {
			return nil, err
		}
		all = append(all, r.Standings.Results...)
		if !r.Standings.HasNext || len(r.Standings.Results) == 0 {
			break
		}
		page++
	}
	if len(all) > maxManagers {
		all = all[:maxManagers]
	}
	return all, nil
}

func (c *client) fetchEventLive(ctx context.Context, gw int) (*eventLiveResponse, error) {
	var r eventLiveResponse
	return &r, c.get(ctx, fmt.Sprintf("/event/%d/live/", gw), &r)
}

func (c *client) fetchPicks(ctx context.Context, managerID, gw int) (*picksResponse, error) {
	var r picksResponse
	path := fmt.Sprintf("/entry/%d/event/%d/picks/", managerID, gw)
	return &r, c.get(ctx, path, &r)
}

// entryResponse is the subset of /entry/{id}/ we use. Bank and value are in
// tenths of a million (e.g. 1005 = £100.5m). LastDeadlineValue is the total
// squad selling value plus bank at the last deadline.
type entryResponse struct {
	LastDeadlineBank  int `json:"last_deadline_bank"`
	LastDeadlineValue int `json:"last_deadline_value"`
}

func (c *client) fetchEntry(ctx context.Context, managerID int) (*entryResponse, error) {
	var r entryResponse
	return &r, c.get(ctx, fmt.Sprintf("/entry/%d/", managerID), &r)
}
