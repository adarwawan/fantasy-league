package wcf

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	baseURL         = "https://play.fifa.com"
	managersPerPage = 20
)

type client struct {
	http   *http.Client
	cookie string // WC_COOKIE value for authenticated endpoints
}

func newClient(cookie string) *client {
	return &client{
		http:   &http.Client{Timeout: 15 * time.Second},
		cookie: cookie,
	}
}

func (c *client) get(ctx context.Context, path string, dst any, auth bool) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "fantasy-league-dashboard/0.1")
	if auth && c.cookie != "" {
		req.Header.Set("Cookie", c.cookie)
	}

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

type wcfPlayer struct {
	ID              int            `json:"id"`
	FirstName       string         `json:"firstName"`
	LastName        string         `json:"lastName"`
	KnownName       *string        `json:"knownName"`
	SquadID         int            `json:"squadId"`
	Position        string         `json:"position"`
	Price           float64        `json:"price"`
	Status          string         `json:"status"` // "playing" | "suspended" | "transferred"
	PercentSelected float64        `json:"percentSelected"`
	Stats           wcfPlayerStats `json:"stats"`
}

type wcfPlayerStats struct {
	TotalPoints     int     `json:"totalPoints"`
	LastRoundPoints int     `json:"lastRoundPoints"`
	Form            float64 `json:"form"`
	AvgPoints       float64 `json:"avgPoints"`
}

type wcfSquad struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Abbr string `json:"abbr"`
}

type wcfRound struct {
	ID          int             `json:"id"`
	Status      string          `json:"status"`
	StartDate   string          `json:"startDate"`
	EndDate     string          `json:"endDate"`
	Tournaments []wcfTournament `json:"tournaments"`
	Stage       string          `json:"stage"`
}

type wcfTournament struct {
	ID            int    `json:"id"`
	Period        string `json:"period"`
	Minutes       int    `json:"minutes"`
	ExtraMinutes  int    `json:"extraMinutes"`
	VenueName     string `json:"venueName"`
	VenueCity     string `json:"venueCity"`
	VenueID       int    `json:"venueId"`
	Date          string `json:"date"`
	Status        string `json:"status"`
	IsSuspended   bool   `json:"isSuspended"`
	HomeSquadID   int    `json:"homeSquadId"`
	AwaySquadID   int    `json:"awaySquadId"`
	HomeSquadName string `json:"homeSquadName"`
	AwaySquadName string `json:"awaySquadName"`
	HomeSquadAbbr string `json:"homeSquadAbbr"`
	AwaySquadAbbr string `json:"awaySquadAbbr"`
	HomeScore     *int   `json:"homeScore"`
	AwayScore     *int   `json:"awayScore"`
	HomePenalty   *int   `json:"homePenaltyScore"`
	AwayPenalty   *int   `json:"awayPenaltyScore"`
}

type rankingResponse struct {
	Success struct {
		Ranks []wcfRankEntry `json:"ranks"`
	} `json:"success"`
}

type wcfRankEntry struct {
	UserID        int64  `json:"userId"`
	UserName      string `json:"userName"`
	LeagueID      int    `json:"leagueId"`
	RoundID       int    `json:"roundId"`
	RoundRank     int    `json:"roundRank"`
	RoundPoints   int    `json:"roundPoints"`
	OverallRank   int    `json:"overallRank"`
	OverallPoints int    `json:"overallPoints"`
	Level         int    `json:"level"`
	Avatar        string `json:"avatar"`
}

type historyResponse struct {
	Success wcfPickEntry `json:"success"`
}

type wcfPickEntry struct {
	Captain           *int             `json:"captain"`
	Vice              *int             `json:"vice"`
	Lineup            map[string][]int `json:"lineup"`
	Bench             map[string][]int `json:"bench"`
	BenchOrder        []int            `json:"benchOrder"`
	WildCard          interface{}      `json:"wildCard"`
	FreeHit           interface{}      `json:"freeHit"`
	TwelfthMan        interface{}      `json:"twelfthMan"`
	MaxCaptainBooster interface{}      `json:"maxCaptainBooster"`
}

func (c *client) fetchPlayers(ctx context.Context) ([]wcfPlayer, error) {
	var r []wcfPlayer
	return r, c.get(ctx, "/json/fantasy/players.json", &r, false)
}

func (c *client) fetchSquads(ctx context.Context) ([]wcfSquad, error) {
	var r []wcfSquad
	return r, c.get(ctx, "/json/fantasy/squads.json", &r, false)
}

func (c *client) fetchRounds(ctx context.Context) ([]wcfRound, error) {
	var r []wcfRound
	return r, c.get(ctx, "/json/fantasy/rounds.json", &r, false)
}

func (c *client) fetchRankings(ctx context.Context, maxManagers int) ([]wcfRankEntry, error) {
	var all []wcfRankEntry
	page := 1
	for len(all) < maxManagers {
		var r rankingResponse
		path := fmt.Sprintf("/api/en/fantasy/ranking/overall?page=%d&limit=%d", page, managersPerPage)
		if err := c.get(ctx, path, &r, true); err != nil {
			return nil, err
		}
		all = append(all, r.Success.Ranks...)
		if len(r.Success.Ranks) < managersPerPage {
			break
		}
		page++
	}
	if len(all) > maxManagers {
		all = all[:maxManagers]
	}
	return all, nil
}

func (c *client) fetchHistory(ctx context.Context, gw, managerID int) (*historyResponse, error) {
	var r historyResponse
	path := fmt.Sprintf("/api/en/fantasy/team/history/%d/%d", gw, managerID)
	return &r, c.get(ctx, path, &r, true)
}
