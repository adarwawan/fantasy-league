// Package trends is an isolated FPL transfer-velocity service. It polls player
// ownership/transfer counts every few minutes during the ~24h before a deadline
// and serves ownership-over-time and velocity ("fastest movers") views.
//
// Like internal/setpiece it owns only its own tables (trends_*) and shares the
// pgx pool + Redis cache, but it is wired through its own binary (cmd/trends)
// on a separate port rather than the main server.
package trends

import "time"

// Direction selects which end of the velocity spectrum the leaders board ranks:
// inflows (transfers in surging) or outflows (mass exodus).
type Direction string

const (
	DirectionIn  Direction = "in"
	DirectionOut Direction = "out"
)

// ParseDirection maps a query value to a Direction, defaulting to "in".
func ParseDirection(s string) Direction {
	if s == string(DirectionOut) {
		return DirectionOut
	}
	return DirectionIn
}

// Metric selects what "velocity" measures. Transfers (net in-minus-out delta)
// is the sharper signal, but it is meaningless in the GW1 window because initial
// squad selection isn't a transfer — the counters sit at ~0 while ownership
// moves. So GW1 falls back to ownership (net selected_by_percent delta).
type Metric string

const (
	MetricTransfers Metric = "transfers"
	MetricOwnership Metric = "ownership"
)

// MetricForGameweek picks the ranking metric for a window automatically: GW1
// uses ownership (no transfers exist before the first deadline); GW2+ uses net
// transfers.
func MetricForGameweek(gw int) Metric {
	if gw <= 1 {
		return MetricOwnership
	}
	return MetricTransfers
}

// Session is an armed capture window for one gameweek.
type Session struct {
	Gameweek  int       `json:"gameweek"`
	StartedAt time.Time `json:"started_at"`
	EndsAt    time.Time `json:"ends_at"`
	Active    bool      `json:"active"`
	PollCount int       `json:"poll_count"`
}

// Snapshot is one player's captured state at one poll.
type Snapshot struct {
	PlayerExtID  int
	SelectedBP   int // selected_by_percent in basis points (1234 = 12.34%)
	TransfersIn  int // transfers_in_event  (cumulative this GW)
	TransfersOut int // transfers_out_event (cumulative this GW)
	NowCost      int // price in tenths
}

// NetTransfers is the running in-minus-out for this GW.
func (s Snapshot) NetTransfers() int { return s.TransfersIn - s.TransfersOut }

// LeaderRow is one entry on the "Fastest Movers" board.
type LeaderRow struct {
	PlayerExtID  int     `json:"player_ext_id"`
	Name         string  `json:"name"`
	Team         string  `json:"team"`
	Position     string  `json:"position"`
	SelectedPct  float64 `json:"selected_pct"`
	NowCost      float64 `json:"now_cost"` // in millions
	NetTransfers int     `json:"net_transfers"`
	// RankDelta is the change over the trailing window that the board ranks by —
	// the velocity signal. Its unit follows the active metric: an integer transfer
	// count (transfers) or basis points of ownership (ownership).
	RankDelta int `json:"rank_delta"`
}

// SeriesPoint is one sample in a player's snapshot series.
type SeriesPoint struct {
	CapturedAt   time.Time `json:"captured_at"`
	SelectedPct  float64   `json:"selected_pct"`
	TransfersIn  int       `json:"transfers_in"`
	TransfersOut int       `json:"transfers_out"`
	NetTransfers int       `json:"net_transfers"`
	NowCost      float64   `json:"now_cost"`
}
