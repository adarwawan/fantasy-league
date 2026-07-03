package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"fantasy-league/internal/musthave"
	"fantasy-league/internal/store"
)

type playerStore interface {
	QueryPlayers(ctx context.Context, gameID, pos string, maxPrice float64, sort string, topN int) ([]store.PlayerRow, error)
	QueryPlayerOwnerships(ctx context.Context, gameID string) ([]store.PlayerOwnership, error)
	QueryRecentGWPoints(ctx context.Context, gameID string, window int) (map[string][]int, int, error)
	CurrentGW(ctx context.Context, gameID string) (int, error)
}

type cacheStore interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
}

// PlayersHandler handles GET /api/{game}/players and GET /api/{game}/players/scatter.
type PlayersHandler struct {
	store    playerStore
	cache    cacheStore
	mustHave map[string]musthave.Config
}

// NewPlayersHandler creates the handler. mustHave holds per-game must-have
// thresholds; games without an entry fall back to musthave.DefaultConfig.
func NewPlayersHandler(s playerStore, c cacheStore, mustHave map[string]musthave.Config) *PlayersHandler {
	return &PlayersHandler{store: s, cache: c, mustHave: mustHave}
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

	pos := q.Get("pos")
	sortBy := q.Get("sort")
	if sortBy == "" {
		sortBy = "global_ownership"
	}
	maxPrice, _ := strconv.ParseFloat(q.Get("max_price"), 64)
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

	resp := buildPlayersResponse(game, gw, topN, players, h.mustHaveFlags(r.Context(), game, gw, players))
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

	resp := buildPlayersResponse(game, gw, scatterTopN, players, h.mustHaveFlags(r.Context(), game, gw, players))
	b, _ := json.Marshal(resp)
	h.cache.Set(r.Context(), cacheKey, b, 30*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(b)
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

// mustHaveFlags computes the must-have flag for each player. Stars are
// auxiliary, so lookup failures degrade to no flags rather than failing the
// request.
func (h *PlayersHandler) mustHaveFlags(ctx context.Context, game string, gw int, players []store.PlayerRow) map[string]bool {
	cfg, ok := h.mustHave[game]
	if !ok {
		cfg = musthave.DefaultConfig()
	}

	pool, err := h.store.QueryPlayerOwnerships(ctx, game)
	if err != nil {
		slog.Warn("must-have: query ownerships failed", "game", game, "err", err)
		return nil
	}
	points, gwsCounted, err := h.store.QueryRecentGWPoints(ctx, game, cfg.FormWindow)
	if err != nil {
		slog.Warn("must-have: query recent points failed", "game", game, "err", err)
		return nil
	}
	return musthave.Compute(players, pool, points, gwsCounted, gw, cfg)
}

func buildPlayersResponse(game string, gw, topN int, rows []store.PlayerRow, mustHave map[string]bool) playersResponse {
	players := make([]playerJSON, len(rows))
	for i, r := range rows {
		fixtures := make([]fixtureJSON, len(r.Fixtures))
		for j, f := range r.Fixtures {
			fixtures[j] = fixtureJSON{
				GW: f.GW, Opp: f.Opp, HA: f.HA,
				Difficulty: f.Difficulty, Kickoff: f.Kickoff,
			}
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
