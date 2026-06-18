package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"fantasy-league/internal/store"
)

type playerStore interface {
	QueryPlayers(ctx context.Context, gameID, pos string, maxPrice float64, sort string) ([]store.PlayerRow, error)
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
	TopNSize        int           `json:"top_n_size"`
	Status          string        `json:"status"`
	News            string        `json:"news"`
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
	if topN == 0 {
		topN = 10000
	}

	cacheKey := store.CacheKey(game, "players", pos, sortBy, strconv.Itoa(topN), strconv.FormatFloat(maxPrice, 'f', 1, 64))
	if cached, _ := h.cache.Get(r.Context(), cacheKey); cached != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write(cached)
		return
	}

	players, err := h.store.QueryPlayers(r.Context(), game, pos, maxPrice, sortBy)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	gw, _ := h.store.CurrentGW(r.Context(), game)

	resp := buildPlayersResponse(game, gw, topN, players)
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

	players, err := h.store.QueryPlayers(r.Context(), game, "", 0, "global_ownership")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	gw, _ := h.store.CurrentGW(r.Context(), game)

	resp := buildPlayersResponse(game, gw, 10000, players)
	b, _ := json.Marshal(resp)
	h.cache.Set(r.Context(), cacheKey, b, 30*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(b)
}

func buildPlayersResponse(game string, gw, topN int, rows []store.PlayerRow) playersResponse {
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
			TopNSize:        r.TopNSize,
			Status:          r.Status,
			News:            r.News,
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
