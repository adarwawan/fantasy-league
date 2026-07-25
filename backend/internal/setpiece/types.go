package setpiece

import "time"

// This package is deliberately isolated: it never imports or mutates the main
// app's player/fixture tables (see docs/set-piece-detector.md §3). Player
// identity is keyed by Understat id throughout; teams are grouped by a canonical
// PL name resolved via a small override map.

// Role distinguishes the two observed signals mined from the same shot stream.
type Role string

const (
	// RoleTaker is the delivery signal: the shooter *is* the set-piece taker
	// (penalties, direct free-kicks).
	RoleTaker Role = "taker"
	// RoleTarget is the receiving signal: the shooter got on the end of an
	// indirect set piece (corners, set-piece free-kicks) — the target man.
	RoleTarget Role = "target"
)

// Duty is the specific set-piece type within a role.
type Duty string

const (
	DutyPenalty  Duty = "penalty"  // taker
	DutyDFK      Duty = "dfk"      // taker: direct free-kick
	DutyCorner   Duty = "corner"   // target
	DutySetPiece Duty = "setpiece" // target: indirect free-kick delivery
	// DutyAll is the cross-duty aggregate row for target men, used for the
	// league-wide ranking. Takers never use it.
	DutyAll Duty = "all"
)

// Understat `situation` values we classify on.
const (
	situationPenalty        = "Penalty"
	situationDirectFreekick = "DirectFreekick"
	situationFromCorner     = "FromCorner"
	situationSetPiece       = "SetPiece"
)

// --- Understat JSON wire types ---------------------------------------------
//
// Understat encodes every numeric field as a string. Decode into strings and
// convert in parse.go.

// leagueData is the getLeagueData/EPL/{season} envelope; only `dates` (the
// fixture list) is needed.
type leagueData struct {
	Dates []leagueMatch `json:"dates"`
}

// leagueMatch is one entry from getLeagueData's `dates` array.
type leagueMatch struct {
	ID       string     `json:"id"`
	IsResult bool       `json:"isResult"`
	H        leagueTeam `json:"h"`
	A        leagueTeam `json:"a"`
	Datetime string     `json:"datetime"`
}

type leagueTeam struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	ShortTitle string `json:"short_title"`
}

// matchData is the payload from getMatchData/{match_id}. Only `shots` is needed
// for the detector; `rosters`/`match_info` are ignored.
type matchData struct {
	Shots matchShots `json:"shots"`
}

type matchShots struct {
	H []shot `json:"h"`
	A []shot `json:"a"`
}

// shot is one Understat shot object. Only the fields the detector needs are
// decoded; the rest (result, X, Y, xGChain, …) are left out.
type shot struct {
	Minute    string `json:"minute"`
	Situation string `json:"situation"`
	ShotType  string `json:"shotType"`
	Player    string `json:"player"`
	PlayerID  string `json:"player_id"`
	HA        string `json:"h_a"`
	HTeam     string `json:"h_team"`
	ATeam     string `json:"a_team"`
	Season    string `json:"season"`
	Date      string `json:"date"`
	XG        string `json:"xG"`
	Result    string `json:"result"` // "Goal" | "SavedShot" | "MissedShots" | ...
}

// --- Internal domain types --------------------------------------------------

// Event is a single qualifying set-piece shot, the subject being the shooter.
// One shot yields at most one Event (it matches exactly one role/duty).
type Event struct {
	MatchID       string
	Season        string
	MatchDate     time.Time
	Minute        int
	UnderstatTeam string // canonical PL team name
	Role          Role
	Duty          Duty
	PlayerID      string
	PlayerName    string
	IsHeader      bool
	IsGoal        bool
	XG            float64
}

// BoardRow is one ranked entry over the rolling window for a team/role/duty.
type BoardRow struct {
	UnderstatTeam string
	Role          Role
	Duty          Duty
	PlayerID      string
	PlayerName    string
	Rank          int
	WeightedScore float64
	RawCount      int
	Goals         int
	XGSum         float64
	HeaderPct     *float64 // target-man context; nil for takers
	LastSeen      time.Time
}
