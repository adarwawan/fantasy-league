package handler

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/go-chi/chi/v5"

	"fantasy-league/internal/store"
)

// bonusTopPlayers is how many bonus earners each club lists. Always three, so
// the frontend can derive a top-3 concentration share from the same rows.
const bonusTopPlayers = 3

type teamBonusPlayer struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Position string `json:"position"`
	Bonus    int    `json:"bonus"`
}

// teamBonusByPosition is the club's bonus split by the position line that
// earned it. Raw point totals; shares and per-match rates are derived client-side.
type teamBonusByPosition struct {
	GK  int `json:"GK"`
	DEF int `json:"DEF"`
	MID int `json:"MID"`
	FWD int `json:"FWD"`
}

type teamBonusEntry struct {
	Team       string              `json:"team"`    // short name, e.g. "ARS"
	Matches    int                 `json:"matches"` // finished fixtures in the window
	Bonus      int                 `json:"bonus"`
	ByPosition teamBonusByPosition `json:"by_position"`
	Top        []teamBonusPlayer   `json:"top"`
}

type teamBonusResponse struct {
	Teams []teamBonusEntry `json:"teams"`
	Meta  statsMetaJSON    `json:"meta"`
}

// computeTeamBonus attributes every bonus point in the window to the player's
// current club, then splits each club's haul by position and lists its top
// earners. `matches` is finished fixtures per team over the same window; clubs
// with no finished match are dropped so the frontend never divides by zero.
// Bonus is attributed to the current club, so a mid-window mover's bonus all
// lands on the new club (same caveat as the rest of the stats views).
// Teams are ordered by short name, earners by bonus descending (ties: name).
func computeTeamBonus(lines []store.PlayerStatGW, matches map[string]int, limit int) []teamBonusEntry {
	type agg struct {
		name, position string
		bonus          int
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
		a.bonus += l.Bonus
	}

	teams := make([]string, 0, len(matches))
	for t, n := range matches {
		if n > 0 {
			teams = append(teams, t)
		}
	}
	sort.Strings(teams)

	out := make([]teamBonusEntry, 0, len(teams))
	for _, t := range teams {
		entry := teamBonusEntry{Team: t, Matches: matches[t], Top: []teamBonusPlayer{}}
		var players []teamBonusPlayer
		for id, a := range byTeam[t] {
			entry.Bonus += a.bonus
			switch a.position {
			case "GK":
				entry.ByPosition.GK += a.bonus
			case "DEF":
				entry.ByPosition.DEF += a.bonus
			case "MID":
				entry.ByPosition.MID += a.bonus
			case "FWD":
				entry.ByPosition.FWD += a.bonus
			}
			if a.bonus > 0 {
				players = append(players, teamBonusPlayer{ID: id, Name: a.name, Position: a.position, Bonus: a.bonus})
			}
		}
		sort.Slice(players, func(i, j int) bool {
			if players[i].Bonus != players[j].Bonus {
				return players[i].Bonus > players[j].Bonus
			}
			return players[i].Name < players[j].Name
		})
		if len(players) > limit {
			players = players[:limit]
		}
		if players != nil {
			entry.Top = players
		}
		out = append(out, entry)
	}
	return out
}

// Bonus handles GET /api/{game}/stats/bonus: per-club bonus totals, position
// split and top earners over the same finished-GW window as the other stats.
func (h *StatsHandler) Bonus(w http.ResponseWriter, r *http.Request) {
	game := chi.URLParam(r, "game")
	cacheKey := store.CacheKey(game, "stats", "bonus")

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
	matches, err := h.store.QueryRecentTeamMatches(r.Context(), game, statsWindow)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	gw, _ := h.store.CurrentGW(r.Context(), game)

	resp := teamBonusResponse{
		Teams: computeTeamBonus(lines, matches, bonusTopPlayers),
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
