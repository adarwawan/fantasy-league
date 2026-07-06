package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/go-chi/chi/v5"

	"fantasy-league/internal/store"
)

// statsWindow is how many finished gameweeks the Stats page aggregates over.
const statsWindow = 5

// statsLeaderLimit is how many players each stat card lists.
const statsLeaderLimit = 5

// Defensive-contribution thresholds (FPL scoring rules): a player earns 2 points
// in a gameweek when a defender reaches 10+ CBIT actions, or a midfielder/forward
// reaches 12+ CBIT+recovery actions.
const (
	defenderDCThreshold = 10
	midFwdDCThreshold   = 12
	defensiveConPoints  = 2
)

type statsStore interface {
	QueryRecentPlayerStatLines(ctx context.Context, gameID string, window int) ([]store.PlayerStatGW, error)
	CurrentGW(ctx context.Context, gameID string) (int, error)
}

// statLeader is a player's aggregated total for one scoring component within a
// position, with a 1-based rank (top rank = highest total). It is the service
// layer's output, consumed by buildStatsResponse.
type statLeader struct {
	Position  string
	Component string
	Rank      int
	PlayerID  string
	Name      string
	Team      string
	Value     int
}

// defensiveContributionPoints returns the FPL points a player earns in a single
// gameweek for defensive contribution. Pure function so the scoring rule can be
// unit-tested directly.
func defensiveContributionPoints(position string, actions int) int {
	switch position {
	case "DEF":
		if actions >= defenderDCThreshold {
			return defensiveConPoints
		}
	case "MID", "FWD":
		if actions >= midFwdDCThreshold {
			return defensiveConPoints
		}
	}
	return 0
}

// componentPoints converts a player's raw count for a counting component
// (goals, assists, clean sheets) into the FPL points it earns, which depends on
// position. Bonus and defensive contribution are already point totals and do not
// pass through here. Positions/components with no scoring (e.g. FWD clean sheet)
// return 0. Pure function so the scoring rule is unit-testable.
//
// Point values (https://fantasy.premierleague.com/help/rules):
//   goals       — GK/DEF 6, MID 5, FWD 4
//   assists     — 3 for all positions
//   clean sheet — GK/DEF 4, MID 1, FWD 0
func componentPoints(position, component string, count int) int {
	switch component {
	case "goals":
		switch position {
		case "GK", "DEF":
			return count * 6
		case "MID":
			return count * 5
		case "FWD":
			return count * 4
		}
	case "assists":
		return count * 3
	case "clean_sheet":
		switch position {
		case "GK", "DEF":
			return count * 4
		case "MID":
			return count * 1
		}
	}
	return 0
}

// statComponents is the fixed set of scoring components ranked per position.
var statComponents = []string{"goals", "assists", "clean_sheet", "bonus", "defensive_con"}

// pointScoredComponents are the counting components whose per-player totals are
// converted to points via componentPoints. "bonus" and "defensive_con" are
// already accumulated as points, so they are excluded.
var pointScoredComponents = map[string]bool{"goals": true, "assists": true, "clean_sheet": true}

// statPositions is the fixed set of positions leaders are ranked within.
var statPositions = []string{"GK", "DEF", "MID", "FWD"}

// computeStatLeaders aggregates raw per-gameweek stat lines into ranked leaders:
// for each position and component, the top `limit` players by FPL points earned,
// excluding zero totals. Every card's value is points: goals/assists/clean sheets
// are converted from raw counts via componentPoints; "bonus" is already points;
// "defensive_con" sums the per-GW threshold points (defensiveContributionPoints).
// Deterministic: ties break on player name ascending.
func computeStatLeaders(lines []store.PlayerStatGW, limit int) []statLeader {
	// Aggregate per player. component -> total.
	type agg struct {
		position, name, team string
		totals               map[string]int
	}
	byPlayer := make(map[string]*agg)
	order := make([]string, 0) // stable set of player IDs (unused for output order, ranking sorts)
	for _, l := range lines {
		a := byPlayer[l.PlayerID]
		if a == nil {
			a = &agg{position: l.Position, name: l.Name, team: l.TeamShortName, totals: map[string]int{}}
			byPlayer[l.PlayerID] = a
			order = append(order, l.PlayerID)
		}
		a.totals["goals"] += l.Goals
		a.totals["assists"] += l.Assists
		a.totals["clean_sheet"] += l.CleanSheets
		a.totals["bonus"] += l.Bonus
		a.totals["defensive_con"] += defensiveContributionPoints(l.Position, l.DefensiveContribution)
	}

	var out []statLeader
	for _, pos := range statPositions {
		for _, comp := range statComponents {
			type cand struct {
				id, name, team string
				value          int
			}
			var cands []cand
			for _, id := range order {
				a := byPlayer[id]
				if a.position != pos {
					continue
				}
				v := a.totals[comp]
				if pointScoredComponents[comp] {
					v = componentPoints(a.position, comp, v)
				}
				if v > 0 {
					cands = append(cands, cand{id, a.name, a.team, v})
				}
			}
			sort.Slice(cands, func(i, j int) bool {
				if cands[i].value != cands[j].value {
					return cands[i].value > cands[j].value
				}
				return cands[i].name < cands[j].name
			})
			for i, c := range cands {
				if i >= limit {
					break
				}
				out = append(out, statLeader{
					Position: pos, Component: comp, Rank: i + 1,
					PlayerID: c.id, Name: c.name, Team: c.team, Value: c.value,
				})
			}
		}
	}
	return out
}

// StatsHandler handles GET /api/{game}/stats.
type StatsHandler struct {
	store statsStore
	cache cacheStore
}

func NewStatsHandler(s statsStore, c cacheStore) *StatsHandler {
	return &StatsHandler{store: s, cache: c}
}

// statCardSpec defines one card: which component to pull and how to present it.
type statCardSpec struct {
	Component string // matches statLeader.Component
	Label     string
	Points    string // FPL point value, e.g. "+4" or "1-3"
}

// positionCards is the per-position ordering of stat cards, mirroring the FPL
// scoring rules (https://fantasy.premierleague.com/help/rules).
var positionCards = []struct {
	Position string
	Label    string
	Cards    []statCardSpec
}{
	{"GK", "Goalkeeper", []statCardSpec{
		{"clean_sheet", "Clean Sheet", "+4"},
		{"bonus", "Bonus", "1-3"},
	}},
	{"DEF", "Defender", []statCardSpec{
		{"goals", "Goals Scored", "+6"},
		{"clean_sheet", "Clean Sheet", "+4"},
		{"assists", "Assists", "+3"},
		{"bonus", "Bonus", "1-3"},
		{"defensive_con", "Defensive Contribution", "+2"},
	}},
	{"MID", "Midfielder", []statCardSpec{
		{"goals", "Goals Scored", "+5"},
		{"assists", "Assists", "+3"},
		{"defensive_con", "Defensive Contribution", "+2"},
		{"bonus", "Bonus", "1-3"},
	}},
	{"FWD", "Forward", []statCardSpec{
		{"goals", "Goals Scored", "+4"},
		{"assists", "Assists", "+3"},
		{"bonus", "Bonus", "1-3"},
	}},
}

type statLeaderJSON struct {
	Rank  int    `json:"rank"`
	ID    string `json:"id"`
	Name  string `json:"name"`
	Team  string `json:"team"`
	Value int    `json:"value"`
}

type statCardJSON struct {
	Component string           `json:"component"`
	Label     string           `json:"label"`
	Points    string           `json:"points"`
	Leaders   []statLeaderJSON `json:"leaders"`
}

type statSectionJSON struct {
	Position string         `json:"position"`
	Label    string         `json:"label"`
	Cards    []statCardJSON `json:"cards"`
}

type statsMetaJSON struct {
	GameID   string    `json:"game_id"`
	GW       int       `json:"gw"`
	Window   int       `json:"window"`
	CachedAt time.Time `json:"cached_at"`
}

type statsResponse struct {
	Sections []statSectionJSON `json:"sections"`
	Meta     statsMetaJSON     `json:"meta"`
}

func (h *StatsHandler) List(w http.ResponseWriter, r *http.Request) {
	game := chi.URLParam(r, "game")
	cacheKey := store.CacheKey(game, "stats")

	if cached, _ := h.cache.Get(r.Context(), cacheKey); cached != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write(cached)
		return
	}

	lines, err := h.store.QueryRecentPlayerStatLines(r.Context(), game, statsWindow)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	gw, _ := h.store.CurrentGW(r.Context(), game)

	leaders := computeStatLeaders(lines, statsLeaderLimit)
	resp := buildStatsResponse(game, gw, leaders)
	b, _ := json.Marshal(resp)
	h.cache.Set(r.Context(), cacheKey, b, 30*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(b)
}

// buildStatsResponse groups the flat leader rows into the position/component
// card layout defined by positionCards. Only components present in the spec are
// emitted; a card with no leaders is still included (empty list) so the page
// layout is stable.
func buildStatsResponse(game string, gw int, leaders []statLeader) statsResponse {
	// index[position][component] -> ordered leaders
	index := make(map[string]map[string][]statLeaderJSON)
	for _, r := range leaders {
		if index[r.Position] == nil {
			index[r.Position] = make(map[string][]statLeaderJSON)
		}
		index[r.Position][r.Component] = append(index[r.Position][r.Component], statLeaderJSON{
			Rank: r.Rank, ID: r.PlayerID, Name: r.Name, Team: r.Team, Value: r.Value,
		})
	}

	sections := make([]statSectionJSON, 0, len(positionCards))
	for _, pc := range positionCards {
		cards := make([]statCardJSON, 0, len(pc.Cards))
		for _, spec := range pc.Cards {
			leaders := index[pc.Position][spec.Component]
			if leaders == nil {
				leaders = []statLeaderJSON{}
			}
			cards = append(cards, statCardJSON{
				Component: spec.Component,
				Label:     spec.Label,
				Points:    spec.Points,
				Leaders:   leaders,
			})
		}
		sections = append(sections, statSectionJSON{
			Position: pc.Position,
			Label:    pc.Label,
			Cards:    cards,
		})
	}

	return statsResponse{
		Sections: sections,
		Meta: statsMetaJSON{
			GameID:   game,
			GW:       gw,
			Window:   statsWindow,
			CachedAt: time.Now().UTC(),
		},
	}
}
