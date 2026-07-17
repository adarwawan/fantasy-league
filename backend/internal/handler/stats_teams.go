package handler

import (
	"encoding/json"
	"math"
	"net/http"
	"sort"
	"time"

	"github.com/go-chi/chi/v5"

	"fantasy-league/internal/store"
)

// teamICTLeaderLimit is how many players each team card lists.
const teamICTLeaderLimit = 5

// teamICTPlayer is one player's ICT totals over the stats window and his share
// of the team's total ICT. Share is the "focal point" signal: how much of the
// team's underlying attacking involvement flows through this player.
type teamICTPlayer struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Position   string  `json:"position"`
	Influence  float64 `json:"influence"`
	Creativity float64 `json:"creativity"`
	Threat     float64 `json:"threat"`
	ICT        float64 `json:"ict"`
	Share      float64 `json:"share"` // % of team total ICT, 1 decimal
	// Top-3 ranks within the team per component (1-3; omitted when outside
	// the top 3 or the component total is zero). Rendered as star badges.
	InfluenceRank  int `json:"influence_rank,omitempty"`
	CreativityRank int `json:"creativity_rank,omitempty"`
	ThreatRank     int `json:"threat_rank,omitempty"`
}

type teamICTEntry struct {
	Team     string          `json:"team"` // short name, e.g. "ARS"
	TotalICT float64         `json:"total_ict"`
	Players  []teamICTPlayer `json:"players"`
}

type teamICTResponse struct {
	Teams []teamICTEntry `json:"teams"`
	Meta  statsMetaJSON  `json:"meta"`
}

// round1 rounds to 1 decimal, enough for ICT index sums and share percentages.
func round1(v float64) float64 { return math.Round(v*10) / 10 }

// rankTop3 assigns ranks 1-3 for one ICT component across a team's players
// (highest value first, ties broken by name ascending). Zero-value players
// earn no rank. Called before the list is sorted and truncated for display,
// so ranks are against the whole team, not just the emitted top players.
func rankTop3(players []teamICTPlayer, value func(*teamICTPlayer) float64, setRank func(*teamICTPlayer, int)) {
	idx := make([]int, len(players))
	for i := range idx {
		idx[i] = i
	}
	sort.Slice(idx, func(a, b int) bool {
		va, vb := value(&players[idx[a]]), value(&players[idx[b]])
		if va != vb {
			return va > vb
		}
		return players[idx[a]].Name < players[idx[b]].Name
	})
	for r, i := range idx {
		if r >= 3 || value(&players[i]) <= 0 {
			break
		}
		setRank(&players[i], r+1)
	}
}

// computeTeamICTShares aggregates per-GW stat lines into per-team ICT shares:
// for each team, every player's Influence/Creativity/Threat summed over the
// window, his combined ICT, and his share of the team's total ICT. Share is
// computed against the full team total (all players, not just the emitted top
// `limit`), so the listed shares don't sum to 100%. Players with zero ICT are
// dropped; teams are ordered by short name, players by ICT descending
// (ties: name ascending). Badge-holders below the top-`limit` cutoff are
// appended after it so every top-3 component rank is visible on the card.
func computeTeamICTShares(lines []store.PlayerStatGW, limit int) []teamICTEntry {
	type agg struct {
		name, position               string
		influence, creativity, threat float64
	}
	byTeam := make(map[string]map[string]*agg) // team -> playerID -> totals
	for _, l := range lines {
		players := byTeam[l.TeamShortName]
		if players == nil {
			players = make(map[string]*agg)
			byTeam[l.TeamShortName] = players
		}
		a := players[l.PlayerID]
		if a == nil {
			a = &agg{name: l.Name, position: l.Position}
			players[l.PlayerID] = a
		}
		a.influence += l.Influence
		a.creativity += l.Creativity
		a.threat += l.Threat
	}

	teams := make([]string, 0, len(byTeam))
	for t := range byTeam {
		teams = append(teams, t)
	}
	sort.Strings(teams)

	out := make([]teamICTEntry, 0, len(teams))
	for _, t := range teams {
		var total float64
		var players []teamICTPlayer
		for id, a := range byTeam[t] {
			ict := a.influence + a.creativity + a.threat
			total += ict
			if ict > 0 {
				players = append(players, teamICTPlayer{
					ID: id, Name: a.name, Position: a.position,
					Influence:  round1(a.influence),
					Creativity: round1(a.creativity),
					Threat:     round1(a.threat),
					ICT:        round1(ict),
				})
			}
		}
		if total == 0 {
			continue
		}
		rankTop3(players, func(p *teamICTPlayer) float64 { return p.Influence }, func(p *teamICTPlayer, r int) { p.InfluenceRank = r })
		rankTop3(players, func(p *teamICTPlayer) float64 { return p.Creativity }, func(p *teamICTPlayer, r int) { p.CreativityRank = r })
		rankTop3(players, func(p *teamICTPlayer) float64 { return p.Threat }, func(p *teamICTPlayer, r int) { p.ThreatRank = r })
		sort.Slice(players, func(i, j int) bool {
			if players[i].ICT != players[j].ICT {
				return players[i].ICT > players[j].ICT
			}
			return players[i].Name < players[j].Name
		})
		// Keep the top `limit` by combined ICT, plus any player beyond the
		// cutoff who holds a top-3 component badge — a specialist (e.g. the
		// team's #2 threat with modest influence/creativity) stays visible
		// instead of his badge silently vanishing. Extras remain in ICT order
		// after the top block.
		if len(players) > limit {
			kept := players[:limit:limit]
			for _, p := range players[limit:] {
				if p.InfluenceRank+p.CreativityRank+p.ThreatRank > 0 {
					kept = append(kept, p)
				}
			}
			players = kept
		}
		for i := range players {
			players[i].Share = round1(players[i].ICT / total * 100)
		}
		out = append(out, teamICTEntry{Team: t, TotalICT: round1(total), Players: players})
	}
	return out
}

// Teams handles GET /api/{game}/stats/teams: per-team ICT share over the same
// recent-GW window as the position stats.
func (h *StatsHandler) Teams(w http.ResponseWriter, r *http.Request) {
	game := chi.URLParam(r, "game")
	cacheKey := store.CacheKey(game, "stats", "teams")

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

	resp := teamICTResponse{
		Teams: computeTeamICTShares(lines, teamICTLeaderLimit),
		Meta: statsMetaJSON{
			GameID:   game,
			GW:       gw,
			Window:   statsWindow,
			CachedAt: time.Now().UTC(),
		},
	}
	b, _ := json.Marshal(resp)
	h.cache.Set(r.Context(), cacheKey, b, 30*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(b)
}
