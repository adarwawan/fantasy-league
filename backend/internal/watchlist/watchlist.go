// Package watchlist flags differential players: the must-have form, fixture
// and availability rules, but with low (not top-ranked) global ownership.
package watchlist

import (
	"fantasy-league/internal/musthave"
	"fantasy-league/internal/store"
)

// Config is the ownership band, in percent of global ownership, a player must
// sit in: MinOwnership <= ownership < MaxOwnership.
type Config struct {
	MinOwnership float64
	MaxOwnership float64
}

func DefaultConfig() Config {
	return Config{MinOwnership: 2, MaxOwnership: 10}
}

// Compute returns the IDs of watchlist players: those that pass the must-have
// form/fixture/availability rules (mh), sit inside the ownership band, and are
// not already must-have (mustHave wins an overlap). nextGW and the inputs come
// from musthave.Load so both flags share one set of queries.
func Compute(in musthave.Inputs, mh musthave.Config, cfg Config, mustHave map[string]bool) map[string]bool {
	own := make(map[string]float64, len(in.Pool))
	for _, p := range in.Pool {
		own[p.PlayerID] = p.GlobalOwnership
	}
	flags := musthave.Select(in.Players, in.Points, in.GWsCounted, in.NextGW, mh, func(p store.PlayerRow) bool {
		o, ok := own[p.ID]
		return ok && o >= cfg.MinOwnership && o < cfg.MaxOwnership && !mustHave[p.ID]
	})
	if flags == nil {
		return map[string]bool{}
	}
	return flags
}
