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

type teamStore interface {
	QueryTeams(ctx context.Context, gameID string) ([]store.TeamRow, error)
	QueryFixtures(ctx context.Context, gameID string, fromGW, toGW int) ([]store.FixtureRow, error)
}

// TeamsHandler handles GET /api/{game}/teams and GET /api/{game}/fixtures.
type TeamsHandler struct {
	store teamStore
	cache cacheStore
}

func NewTeamsHandler(s teamStore, c cacheStore) *TeamsHandler {
	return &TeamsHandler{store: s, cache: c}
}

type teamResponseItem struct {
	ID        string        `json:"id"`
	GameID    string        `json:"game_id"`
	Name      string        `json:"name"`
	ShortName string        `json:"short_name"`
	AttForm   float64       `json:"att_form"`
	DefForm   float64       `json:"def_form"`
	OvrForm   float64       `json:"ovr_form"`
	Fixtures  []fixtureJSON `json:"fixtures"`
}

type teamsResponse struct {
	Teams []teamResponseItem `json:"teams"`
	Meta  struct {
		GameID   string    `json:"game_id"`
		CachedAt time.Time `json:"cached_at"`
		Total    int       `json:"total"`
	} `json:"meta"`
}

func (h *TeamsHandler) List(w http.ResponseWriter, r *http.Request) {
	game := chi.URLParam(r, "game")
	cacheKey := store.CacheKey(game, "teams")

	if cached, _ := h.cache.Get(r.Context(), cacheKey); cached != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write(cached)
		return
	}

	teams, err := h.store.QueryTeams(r.Context(), game)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}

	items := make([]teamResponseItem, len(teams))
	for i, t := range teams {
		fixtures := make([]fixtureJSON, len(t.Fixtures))
		for j, f := range t.Fixtures {
			fixtures[j] = fixtureJSON{
				GW: f.GW, Opp: f.Opp, HA: f.HA,
				Difficulty: f.Difficulty, Kickoff: f.Kickoff,
			}
		}
		items[i] = teamResponseItem{
			ID:        t.ID,
			GameID:    t.GameID,
			Name:      t.Name,
			ShortName: t.ShortName,
			AttForm:   t.AttForm,
			DefForm:   t.DefForm,
			OvrForm:   t.OvrForm,
			Fixtures:  fixtures,
		}
	}

	resp := teamsResponse{Teams: items}
	resp.Meta.GameID = game
	resp.Meta.CachedAt = time.Now().UTC()
	resp.Meta.Total = len(items)

	b, _ := json.Marshal(resp)
	h.cache.Set(r.Context(), cacheKey, b, 30*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(b)
}

type fixtureResponseItem struct {
	ID             string    `json:"id"`
	GameID         string    `json:"game_id"`
	GW             int       `json:"gw"`
	HomeTeam       string    `json:"home_team"`
	AwayTeam       string    `json:"away_team"`
	HomeDifficulty int       `json:"home_difficulty"`
	AwayDifficulty int       `json:"away_difficulty"`
	KickoffTime    time.Time `json:"kickoff_time"`
	Finished       bool      `json:"finished"`
}

type fixturesResponse struct {
	Fixtures []fixtureResponseItem `json:"fixtures"`
	Meta     struct {
		GameID string `json:"game_id"`
		FromGW int    `json:"from_gw"`
		ToGW   int    `json:"to_gw"`
		Total  int    `json:"total"`
	} `json:"meta"`
}

func (h *TeamsHandler) Fixtures(w http.ResponseWriter, r *http.Request) {
	game := chi.URLParam(r, "game")
	q := r.URL.Query()
	fromGW, _ := strconv.Atoi(q.Get("from_gw"))
	toGW, _ := strconv.Atoi(q.Get("to_gw"))

	cacheKey := store.CacheKey(game, "fixtures", strconv.Itoa(fromGW), strconv.Itoa(toGW))
	if cached, _ := h.cache.Get(r.Context(), cacheKey); cached != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write(cached)
		return
	}

	fixtures, err := h.store.QueryFixtures(r.Context(), game, fromGW, toGW)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}

	items := make([]fixtureResponseItem, len(fixtures))
	for i, f := range fixtures {
		items[i] = fixtureResponseItem{
			ID:             f.ID,
			GameID:         f.GameID,
			GW:             f.GW,
			HomeTeam:       f.HomeShortName,
			AwayTeam:       f.AwayShortName,
			HomeDifficulty: f.HomeDifficulty,
			AwayDifficulty: f.AwayDifficulty,
			KickoffTime:    f.KickoffTime,
			Finished:       f.Finished,
		}
	}

	resp := fixturesResponse{Fixtures: items}
	resp.Meta.GameID = game
	resp.Meta.FromGW = fromGW
	resp.Meta.ToGW = toGW
	resp.Meta.Total = len(items)

	b, _ := json.Marshal(resp)
	h.cache.Set(r.Context(), cacheKey, b, 30*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(b)
}
