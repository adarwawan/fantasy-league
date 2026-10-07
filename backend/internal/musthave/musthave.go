// Package musthave flags "must-have" players: highly owned, in form,
// available, and with a good fixture in their next gameweek.
package musthave

import (
	"context"
	"fmt"
	"math"
	"sort"

	"fantasy-league/internal/store"
)

// Store is the subset of store.Store needed to compute must-have players.
type Store interface {
	QueryPlayers(ctx context.Context, gameID, pos string, maxPrice float64, sort string, topN int) ([]store.PlayerRow, error)
	QueryPlayerOwnerships(ctx context.Context, gameID string) ([]store.PlayerOwnership, error)
	QueryRecentGWPoints(ctx context.Context, gameID string, window int) (map[string][]int, int, error)
	CurrentGW(ctx context.Context, gameID string) (int, error)
}

// Config parameterizes must-have detection. Configs are per game; games
// without one fall back to DefaultConfig.
type Config struct {
	FormWindow    int     // last N finished GWs to inspect
	FormPointsMin int     // GW points threshold that counts as a hit
	FormRatio     float64 // fraction of the inspected GWs that must be hits
	MaxNextFDR    int     // next-GW fixture difficulty cutoff (odds fallback)
	MinXG         float64 // next-GW team xG "green" cutoff (attackers)
	MinCSPct      float64 // next-GW clean-sheet % "green" cutoff (defenders)
	TopGK         int     // ownership rank cutoffs per position
	TopDEF        int
	TopMID        int
	TopFWD        int
}

func DefaultConfig() Config {
	return Config{
		FormWindow:    5,
		FormPointsMin: 6,
		FormRatio:     0.5,
		MaxNextFDR:    3,
		MinXG:         2.0,
		MinCSPct:      50,
		TopGK:         4,
		TopDEF:        8,
		TopMID:        8,
		TopFWD:        5,
	}
}

// ComputeForGame loads a game's inputs from the store and returns the sorted
// IDs of every must-have player. The result is never nil on success, so an
// empty set is distinguishable from a failed computation.
func ComputeForGame(ctx context.Context, s Store, gameID string, cfg Config) ([]string, error) {
	players, err := s.QueryPlayers(ctx, gameID, "", 0, "global_ownership", 0)
	if err != nil {
		return nil, fmt.Errorf("query players: %w", err)
	}
	pool, err := s.QueryPlayerOwnerships(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("query ownerships: %w", err)
	}
	points, gwsCounted, err := s.QueryRecentGWPoints(ctx, gameID, cfg.FormWindow)
	if err != nil {
		return nil, fmt.Errorf("query recent points: %w", err)
	}
	gw, err := s.CurrentGW(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("current gw: %w", err)
	}

	flags := Compute(players, pool, points, gwsCounted, gw, cfg)
	ids := make([]string, 0, len(flags))
	for id := range flags {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids, nil
}

// Compute returns the IDs among candidates that satisfy every must-have
// condition:
//   - global ownership ranks inside the per-position cutoff
//   - scored >= FormPointsMin in at least FormRatio of the gwsCounted
//     inspected GWs (minimum 1 hit)
//   - at least one fixture in their next GW (nextGW, or the following GW if
//     they already played theirs) is "green": for GK/DEF the team's
//     clean-sheet odds >= MinCSPct, for MID/FWD the team's xG >= MinXG; when
//     a fixture carries no odds we fall back to difficulty <= MaxNextFDR.
//     Later GWs don't count (see nextFixtures)
//   - status is available
//
// pool must contain every player in the game so ownership ranks are computed
// over the full pool regardless of any filtering applied to candidates.
// recentPoints maps player ID to points per inspected GW; gwsCounted is how
// many finished GWs were inspected (may be fewer than FormWindow early on).
func Compute(candidates []store.PlayerRow, pool []store.PlayerOwnership, recentPoints map[string][]int, gwsCounted, nextGW int, cfg Config) map[string]bool {
	if nextGW <= 0 {
		return nil
	}

	ranks := ownershipRanks(pool)
	threshold := hitsNeeded(cfg, gwsCounted)
	topByPos := map[string]int{
		"GK":  cfg.TopGK,
		"DEF": cfg.TopDEF,
		"MID": cfg.TopMID,
		"FWD": cfg.TopFWD,
	}

	out := make(map[string]bool)
	for _, p := range candidates {
		if p.Status != "available" {
			continue
		}
		if rank, ok := ranks[p.ID]; !ok || rank > topByPos[p.Position] {
			continue
		}
		if !hasGoodFixture(p.Fixtures, nextGW, p.Position, cfg) {
			continue
		}
		hits := 0
		for _, pts := range recentPoints[p.ID] {
			if pts >= cfg.FormPointsMin {
				hits++
			}
		}
		if hits < threshold {
			continue
		}
		out[p.ID] = true
	}
	return out
}

// hitsNeeded is the qualifying-GW count implied by FormRatio and the number
// of GWs actually inspected, floored at 1 so an empty window never qualifies.
func hitsNeeded(cfg Config, gwsCounted int) int {
	n := int(math.Ceil(cfg.FormRatio * float64(gwsCounted)))
	if n < 1 {
		n = 1
	}
	return n
}

// ownershipRanks assigns each player a 1-based rank within their position by
// global ownership descending, tie-broken by ID for determinism.
func ownershipRanks(pool []store.PlayerOwnership) map[string]int {
	sorted := make([]store.PlayerOwnership, len(pool))
	copy(sorted, pool)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].GlobalOwnership != sorted[j].GlobalOwnership {
			return sorted[i].GlobalOwnership > sorted[j].GlobalOwnership
		}
		return sorted[i].PlayerID < sorted[j].PlayerID
	})

	ranks := make(map[string]int, len(sorted))
	nextRank := map[string]int{}
	for _, p := range sorted {
		nextRank[p.Position]++
		ranks[p.PlayerID] = nextRank[p.Position]
	}
	return ranks
}

// nextFixtures returns the player's fixtures in their next gameweek with a
// match: the earliest GW at or after nextGW. Fixtures holds only unplayed
// games, so a player who already played their nextGW match resolves to the
// following GW, and a double gameweek yields both legs.
func nextFixtures(fixtures []store.FixtureInfo, nextGW int) []store.FixtureInfo {
	target := 0
	for _, f := range fixtures {
		if f.GW >= nextGW && (target == 0 || f.GW < target) {
			target = f.GW
		}
	}
	var out []store.FixtureInfo
	for _, f := range fixtures {
		if target != 0 && f.GW == target {
			out = append(out, f)
		}
	}
	return out
}

// hasGoodFixture reports whether at least one of the player's next-GW fixtures
// is "green". Preference goes to bookmaker odds (fresher than static FDR):
// attackers need team xG >= MinXG, defenders need clean-sheet odds >=
// MinCSPct. When a fixture carries no odds we fall back to difficulty <=
// MaxNextFDR so stars don't vanish whenever odds are absent (e.g. before
// bookmakers price a game, or odds-disabled games).
func hasGoodFixture(fixtures []store.FixtureInfo, nextGW int, pos string, cfg Config) bool {
	for _, f := range nextFixtures(fixtures, nextGW) {
		if green, ok := fixtureIsGreen(f, pos, cfg); ok {
			if green {
				return true
			}
			continue
		}
		if f.Difficulty <= cfg.MaxNextFDR {
			return true
		}
	}
	return false
}

// fixtureIsGreen evaluates a fixture against the odds-based "green" cutoffs for
// the player's position. The second return is false when the fixture lacks the
// odds needed for that position, signalling the caller to fall back to FDR.
func fixtureIsGreen(f store.FixtureInfo, pos string, cfg Config) (green, ok bool) {
	switch pos {
	case "GK", "DEF":
		if f.CSPct == nil {
			return false, false
		}
		return *f.CSPct >= cfg.MinCSPct, true
	default: // MID, FWD
		if f.XG == nil {
			return false, false
		}
		return *f.XG >= cfg.MinXG, true
	}
}
