package shotzone

import "time"

// This package is deliberately isolated, like setpiece: it never imports or
// mutates the main app's player/fixture tables. Player identity is keyed by
// Understat id throughout; teams are grouped by the same canonical PL name
// override map as setpiece.

// Zone is one of the 4 open-play shot zones (see docs: "Shot zone rankings —
// requirements & solution").
type Zone string

const (
	ZoneCenter  Zone = "center"
	ZoneLeft    Zone = "left"
	ZoneRight   Zone = "right"
	ZoneOutside Zone = "outside"
	// ZoneOverall is the cross-zone aggregate, mirroring setpiece's DutyAll.
	ZoneOverall Zone = "overall"
)

// Zones lists the 4 real (non-overall) zones, in the order the frontend
// pitch diagram uses.
var Zones = []Zone{ZoneCenter, ZoneLeft, ZoneRight, ZoneOutside}

// Understat `situation` value we keep — everything else (Penalty, SetPiece,
// DirectFreekick, FromCorner) is a dead-ball shot, tracked instead by the
// setpiece board.
const situationOpenPlay = "OpenPlay"

// Zone geometry thresholds over Understat's normalized 0-1 X/Y shot
// coordinates (attacking goal at X=1). Box-first split: shots outside the
// penalty box are one bucket regardless of width; shots inside the box split
// into thirds by width.
// TODO(open item): confirm these against a sample of real Understat data —
// see the requirements doc's Open items.
const (
	boxLineX     = 0.83
	flankSplitLo = 0.40
	flankSplitHi = 0.60
)

// ClassifyZone buckets a shot by its normalized Understat X/Y coordinates.
func ClassifyZone(x, y float64) Zone {
	if x < boxLineX {
		return ZoneOutside
	}
	switch {
	case y < flankSplitLo:
		return ZoneLeft
	case y < flankSplitHi:
		return ZoneCenter
	default:
		return ZoneRight
	}
}

// --- Understat JSON wire types ---------------------------------------------
//
// A small, self-contained client (client.go) fetches these directly, sharing
// setpiece's Redis cache key for the raw match payload so neither module
// re-scrapes a match the other already fetched.

type leagueData struct {
	Dates []leagueMatch `json:"dates"`
}

type leagueMatch struct {
	ID       string `json:"id"`
	IsResult bool   `json:"isResult"`
}

type matchData struct {
	Shots matchShots `json:"shots"`
}

type matchShots struct {
	H []shot `json:"h"`
	A []shot `json:"a"`
}

// shot is one Understat shot object. Unlike setpiece's detector (which only
// needs the dead-ball fields), this decodes the shot's pitch position too.
type shot struct {
	Minute    string `json:"minute"`
	Situation string `json:"situation"`
	Player    string `json:"player"`
	PlayerID  string `json:"player_id"`
	HA        string `json:"h_a"`
	HTeam     string `json:"h_team"`
	ATeam     string `json:"a_team"`
	Season    string `json:"season"`
	Date      string `json:"date"`
	XG        string `json:"xG"`
	Result    string `json:"result"` // "Goal" | "SavedShot" | "MissedShots" | ...
	X         string `json:"X"`
	Y         string `json:"Y"`
}

// --- Internal domain types --------------------------------------------------

// Event is a single qualifying open-play shot. One shot yields at most one
// Event.
type Event struct {
	MatchID     string
	Season      string
	MatchDate   time.Time
	Minute      int
	TeamFor     string // canonical PL team of the shooter
	TeamAgainst string // canonical PL team defending
	Zone        Zone
	PlayerID    string
	PlayerName  string
	IsGoal      bool
	XG          float64
}

// TakenRow is one ranked player-in-zone entry for a team's own shooters, over
// the rolling window.
type TakenRow struct {
	Team       string
	Zone       Zone
	PlayerID   string
	PlayerName string
	Rank       int
	Shots      int
	Goals      int
	XG         float64
	LastSeen   time.Time
}

// ConcededRow is one team's aggregate shots/goals/xG faced in a zone over the
// rolling window — no player breakdown, a team-weakness signal rather than a
// leaderboard.
type ConcededRow struct {
	Team     string
	Zone     Zone
	Shots    int
	Goals    int
	XG       float64
	LastSeen time.Time
}
