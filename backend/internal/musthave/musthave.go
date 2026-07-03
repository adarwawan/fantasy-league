// Package musthave flags "must-have" players: highly owned, in form,
// available, and with a good fixture in the next gameweek.
package musthave

import (
	"math"
	"sort"

	"fantasy-league/internal/store"
)

// Config parameterizes must-have detection. Configs are per game; games
// without one fall back to DefaultConfig.
type Config struct {
	FormWindow    int     // last N finished GWs to inspect
	FormPointsMin int     // GW points threshold that counts as a hit
	FormRatio     float64 // fraction of the inspected GWs that must be hits
	MaxNextFDR    int     // next-GW fixture difficulty cutoff
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
		TopGK:         4,
		TopDEF:        8,
		TopMID:        8,
		TopFWD:        5,
	}
}

// Compute returns the IDs among candidates that satisfy every must-have
// condition:
//   - global ownership ranks inside the per-position cutoff
//   - scored >= FormPointsMin in at least FormRatio of the gwsCounted
//     inspected GWs (minimum 1 hit)
//   - has a fixture in nextGW with difficulty <= MaxNextFDR
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
		if !hasGoodFixture(p.Fixtures, nextGW, cfg.MaxNextFDR) {
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

// hasGoodFixture reports whether any fixture in nextGW is at or below maxFDR.
func hasGoodFixture(fixtures []store.FixtureInfo, nextGW, maxFDR int) bool {
	for _, f := range fixtures {
		if f.GW == nextGW && f.Difficulty <= maxFDR {
			return true
		}
	}
	return false
}
