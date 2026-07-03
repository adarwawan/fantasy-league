package fpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const baseURL = "https://fantasy.premierleague.com/api"

type client struct {
	http    *http.Client
	baseURL string
}

func newClient() *client {
	return &client{
		http:    &http.Client{Timeout: 15 * time.Second},
		baseURL: baseURL,
	}
}

func (c *client) get(ctx context.Context, path string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "fantasy-league-dashboard/0.1")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: status %d", path, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
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
		TotalPoints int `json:"total_points"`
		GoalsScored int `json:"goals_scored"`
		Assists     int `json:"assists"`
		Bonus       int `json:"bonus"`
	} `json:"stats"`
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

func (c *client) fetchBootstrap(ctx context.Context) (*bootstrapResponse, error) {
	var r bootstrapResponse
	return &r, c.get(ctx, "/bootstrap-static/", &r)
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
