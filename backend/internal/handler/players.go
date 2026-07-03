package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"fantasy-league/internal/store"
)

// recentPointsWindow is how many finished gameweeks of points to show per player.
const recentPointsWindow = 5

type playerStore interface {
	QueryPlayers(ctx context.Context, gameID, pos string, maxPrice float64, sort string, topN int) ([]store.PlayerRow, error)
	QueryRecentGWPointsByGW(ctx context.Context, gameID string, window int) (map[string][]store.GWPoints, error)
	CurrentGW(ctx context.Context, gameID string) (int, error)
}

type cacheStore interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
}

// PlayersHandler handles GET /api/{game}/players and GET /api/{game}/players/scatter.
type PlayersHandler struct {
	store playerStore
	cache cacheStore
}

func NewPlayersHandler(s playerStore, c cacheStore) *PlayersHandler {
	return &PlayersHandler{store: s, cache: c}
}

type fixtureJSON struct {
	GW         int       `json:"gw"`
	Opp        string    `json:"opp"`
	HA         string    `json:"ha"`
	Difficulty int       `json:"difficulty"`
	Kickoff    time.Time `json:"kickoff"`
	XG         *float64  `json:"xg"`
	CSPct      *float64  `json:"cs_pct"`
}

type teamJSON struct {
	ID        string `json:"id"`
	ShortName string `json:"short_name"`
	Name      string `json:"name"`
}

type playerJSON struct {
	ID              string        `json:"id"`
	GameID          string        `json:"game_id"`
	Name            string        `json:"name"`
	Team            teamJSON      `json:"team"`
	Position        string        `json:"position"`
	Price           float64       `json:"price"`
	Form            float64       `json:"form"`
	GlobalOwnership float64       `json:"global_ownership"`
	TopNOwnership   float64       `json:"top_n_ownership"`
	Status          string        `json:"status"`
	News            string        `json:"news"`
	MustHave        bool          `json:"must_have"`
	Fixtures        []fixtureJSON `json:"fixtures"`
	RecentPoints    []gwPointsJSON `json:"recent_points"`
}

type gwPointsJSON struct {
	GW     int `json:"gw"`
	Points int `json:"points"`
}

type metaJSON struct {
	GameID    string    `json:"game_id"`
	GW        int       `json:"gw"`
	TopNSize  int       `json:"top_n_size"`
	CachedAt  time.Time `json:"cached_at"`
	Total     int       `json:"total"`
}

type playersResponse struct {
	Players []playerJSON `json:"players"`
	Meta    metaJSON     `json:"meta"`
}

func (h *PlayersHandler) List(w http.ResponseWriter, r *http.Request) {
	game := chi.URLParam(r, "game")
	q := r.URL.Query()

	pos := canonicalPos(q.Get("pos"))
	sortBy := canonicalPlayerSort(q.Get("sort"))
	rawPrice, _ := strconv.ParseFloat(q.Get("max_price"), 64)
	maxPrice := clampMaxPrice(rawPrice)
	topN, _ := strconv.Atoi(q.Get("top_n"))
	topN = validTopN(game, topN)

	cacheKey := store.CacheKey(game, "players", pos, sortBy, strconv.Itoa(topN), strconv.FormatFloat(maxPrice, 'f', 1, 64))
	if cached, _ := h.cache.Get(r.Context(), cacheKey); cached != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write(cached)
		return
	}

	players, err := h.store.QueryPlayers(r.Context(), game, pos, maxPrice, sortBy, topN)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	gw, _ := h.store.CurrentGW(r.Context(), game)

	resp := buildPlayersResponse(game, gw, topN, players, h.mustHaveFlags(r.Context(), game), h.recentPoints(r.Context(), game))
	b, _ := json.Marshal(resp)
	h.cache.Set(r.Context(), cacheKey, b, 30*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(b)
}

// Scatter returns all players without sort/filter for scatter plot use.
func (h *PlayersHandler) Scatter(w http.ResponseWriter, r *http.Request) {
	game := chi.URLParam(r, "game")
	cacheKey := store.CacheKey(game, "scatter")

	if cached, _ := h.cache.Get(r.Context(), cacheKey); cached != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write(cached)
		return
	}

	scatterTopN := defaultTopN(game)
	players, err := h.store.QueryPlayers(r.Context(), game, "", 0, "global_ownership", scatterTopN)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	gw, _ := h.store.CurrentGW(r.Context(), game)

	resp := buildPlayersResponse(game, gw, scatterTopN, players, h.mustHaveFlags(r.Context(), game), h.recentPoints(r.Context(), game))
	b, _ := json.Marshal(resp)
	h.cache.Set(r.Context(), cacheKey, b, 30*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(b)
}

// maxPlayerPrice bounds the max_price filter so it can't be used to generate
// unlimited distinct cache keys. No game has a player priced above this.
const maxPlayerPrice = 20.0

// validPositions is the finite set of position filters the query accepts. Any
// other value canonicalizes to "" (no filter) so cache keys stay bounded.
var validPositions = map[string]bool{"": true, "GK": true, "DEF": true, "MID": true, "FWD": true}

// canonicalPos normalizes the pos filter to a member of validPositions.
func canonicalPos(pos string) string {
	pos = strings.ToUpper(strings.TrimSpace(pos))
	if !validPositions[pos] {
		return ""
	}
	return pos
}

// validPlayerSorts mirrors the sort keys accepted by store.QueryPlayers. Any
// other value canonicalizes to the default so cache keys stay bounded.
var validPlayerSorts = map[string]bool{
	"global_ownership": true,
	"top_n_ownership":  true,
	"form":             true,
	"price":            true,
	"name":             true,
}

// canonicalPlayerSort normalizes the sort param to a known sort key, defaulting
// to global_ownership. Unknown values already fall back at the query level;
// canonicalizing here keeps distinct sort values from busting the cache.
func canonicalPlayerSort(sort string) string {
	if !validPlayerSorts[sort] {
		return "global_ownership"
	}
	return sort
}

// clampMaxPrice bounds and quantizes the price filter to 0.1 steps within
// [0, maxPlayerPrice]. 0 means "no filter". This caps the number of distinct
// cache keys the filter can produce.
func clampMaxPrice(v float64) float64 {
	if v <= 0 {
		return 0
	}
	if v > maxPlayerPrice {
		v = maxPlayerPrice
	}
	return math.Round(v*10) / 10
}

// topNByGame maps each game to its valid Top-N options (ascending).
var topNByGame = map[string][]int{
	"wcf": {100, 1000},
	"fpl": {1000, 10000, 100000},
}

// validTopN returns n if it is a valid Top-N tier for the game, otherwise the game default.
func validTopN(game string, n int) int {
	opts, ok := topNByGame[game]
	if !ok {
		return n
	}
	for _, o := range opts {
		if o == n {
			return n
		}
	}
	return opts[len(opts)-1] // default to the largest tier
}

// defaultTopN returns the largest Top-N tier for a game.
func defaultTopN(game string) int {
	opts, ok := topNByGame[game]
	if !ok || len(opts) == 0 {
		return 10000
	}
	return opts[len(opts)-1]
}

// mustHaveFlags reads the must-have player IDs computed at sync time from the
// cache. Stars are auxiliary, so a missing or unreadable entry degrades to no
// flags rather than failing the request.
func (h *PlayersHandler) mustHaveFlags(ctx context.Context, game string) map[string]bool {
	b, err := h.cache.Get(ctx, store.CacheKey(game, "musthave"))
	if err != nil || b == nil {
		return nil
	}
	var ids []string
	if err := json.Unmarshal(b, &ids); err != nil {
		slog.Warn("must-have: bad cache entry", "game", game, "err", err)
		return nil
	}
	flags := make(map[string]bool, len(ids))
	for _, id := range ids {
		flags[id] = true
	}
	return flags
}

// recentPoints returns each player's points over the last few finished
// gameweeks. Like must-have stars this is auxiliary, so any error degrades to
// no history rather than failing the request.
func (h *PlayersHandler) recentPoints(ctx context.Context, game string) map[string][]store.GWPoints {
	pts, err := h.store.QueryRecentGWPointsByGW(ctx, game, recentPointsWindow)
	if err != nil {
		slog.Warn("recent points: query failed", "game", game, "err", err)
		return nil
	}
	return pts
}

func buildPlayersResponse(game string, gw, topN int, rows []store.PlayerRow, mustHave map[string]bool, recent map[string][]store.GWPoints) playersResponse {
	players := make([]playerJSON, len(rows))
	for i, r := range rows {
		fixtures := make([]fixtureJSON, len(r.Fixtures))
		for j, f := range r.Fixtures {
			fixtures[j] = fixtureJSON{
				GW: f.GW, Opp: f.Opp, HA: f.HA,
				Difficulty: f.Difficulty, Kickoff: f.Kickoff,
				XG: f.XG, CSPct: f.CSPct,
			}
		}
		recentPts := make([]gwPointsJSON, len(recent[r.ID]))
		for j, gp := range recent[r.ID] {
			recentPts[j] = gwPointsJSON{GW: gp.GW, Points: gp.Points}
		}
		players[i] = playerJSON{
			ID:              r.ID,
			GameID:          r.GameID,
			Name:            r.Name,
			Team:            teamJSON{ID: r.TeamID, ShortName: r.TeamShortName, Name: r.TeamName},
			Position:        r.Position,
			Price:           r.Price,
			Form:            r.Form,
			GlobalOwnership: r.GlobalOwnership,
			TopNOwnership:   r.TopNOwnership,
			Status:          r.Status,
			News:            r.News,
			MustHave:        mustHave[r.ID],
			Fixtures:        fixtures,
			RecentPoints:    recentPts,
		}
	}
	return playersResponse{
		Players: players,
		Meta: metaJSON{
			GameID:   game,
			GW:       gw,
			TopNSize: topN,
			CachedAt: time.Now().UTC(),
			Total:    len(players),
		},
	}
}
