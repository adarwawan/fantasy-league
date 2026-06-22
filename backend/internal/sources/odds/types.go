package odds

import "time"

// GameOddsConfig is injected per game at the call site.
type GameOddsConfig struct {
	GameID      string // "wcf" | "fpl"
	SportKey    string // The Odds API sport key
	StrictSides bool   // false = neutral venue (WCF), true = real home/away (FPL)
}

var WCFOddsConfig = GameOddsConfig{
	GameID:      "wcf",
	SportKey:    "soccer_fifa_world_cup",
	StrictSides: false,
}

var FPLOddsConfig = GameOddsConfig{
	GameID:      "fpl",
	SportKey:    "soccer_epl",
	StrictSides: true,
}

// --- Raw API response types (mirrors the-odds-api.com JSON schema) ---

type OddsMatch struct {
	ID           string      `json:"id"`
	SportKey     string      `json:"sport_key"`
	CommenceTime time.Time   `json:"commence_time"`
	HomeTeam     string      `json:"home_team"`
	AwayTeam     string      `json:"away_team"`
	Bookmakers   []Bookmaker `json:"bookmakers"`
}

type Bookmaker struct {
	Key        string   `json:"key"`
	Title      string   `json:"title"`
	LastUpdate time.Time `json:"last_update"`
	Markets    []Market `json:"markets"`
}

type Market struct {
	Key        string    `json:"key"`
	LastUpdate time.Time `json:"last_update"`
	Outcomes   []Outcome `json:"outcomes"`
}

// Outcome represents one leg of a market.
//
// h2h / h2h_lay: Name is the team name (or "Draw"), Point is nil.
// totals:        Name is "Over" or "Under", Point is the line value (e.g. 2.5).
type Outcome struct {
	Name  string   `json:"name"`
	Price float64  `json:"price"`
	Point *float64 `json:"point,omitempty"`
}

// --- Computed output type ---

// MatchOdds holds computed Poisson estimates for a single match.
type MatchOdds struct {
	OddsMatchID  string
	HomeTeam     string
	AwayTeam     string
	HomeTeamID   string // resolved fantasy.Team.ID; empty if unresolved
	AwayTeamID   string // resolved fantasy.Team.ID; empty if unresolved
	FixtureID    string // linked fantasy.Fixture.ID; empty if unlinked
	KickoffTime  time.Time
	LambdaHome   float64
	LambdaAway   float64
	HomeCSPct    float64 // 0–100
	AwayCSPct    float64 // 0–100
	FetchedAt    time.Time
}
