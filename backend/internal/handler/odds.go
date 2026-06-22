package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"fantasy-league/internal/store"
)

type oddsStore interface {
	QueryMatchOdds(ctx context.Context, gameID string, gws []int, cache *store.Cache) ([]store.MatchOddsRow, error)
	CurrentGW(ctx context.Context, gameID string) (int, error)
}

// OddsHandler handles GET /api/{game}/fixtures/odds.
type OddsHandler struct {
	store oddsStore
	cache *store.Cache
}

func NewOddsHandler(s oddsStore, c *store.Cache) *OddsHandler {
	return &OddsHandler{store: s, cache: c}
}

type oddsResponseItem struct {
	FixtureID   string    `json:"fixture_id,omitempty"`
	GW          int       `json:"gw,omitempty"`
	HomeTeam    string    `json:"home_team"`
	HomeXG      float64   `json:"home_xg"`
	HomeCSPct   float64   `json:"home_cs_pct"`
	AwayTeam    string    `json:"away_team"`
	AwayXG      float64   `json:"away_xg"`
	AwayCSPct   float64   `json:"away_cs_pct"`
	KickoffTime time.Time `json:"kickoff_time"`
}

func (h *OddsHandler) List(w http.ResponseWriter, r *http.Request) {
	game := chi.URLParam(r, "game")
	ctx := r.Context()

	gws := parseGWParam(r.URL.Query().Get("gw"))
	if len(gws) == 0 {
		// Default to current gameweek.
		if cGW, err := h.store.CurrentGW(ctx, game); err == nil && cGW > 0 {
			gws = []int{cGW}
		}
	}

	rows, err := h.store.QueryMatchOdds(ctx, game, gws, h.cache)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}

	items := make([]oddsResponseItem, len(rows))
	for i, row := range rows {
		items[i] = oddsResponseItem{
			FixtureID:   row.FixtureID,
			GW:          row.GW,
			HomeTeam:    row.HomeTeam,
			HomeXG:      row.LambdaHome,
			HomeCSPct:   row.HomeCSPct,
			AwayTeam:    row.AwayTeam,
			AwayXG:      row.LambdaAway,
			AwayCSPct:   row.AwayCSPct,
			KickoffTime: row.KickoffTime,
		}
	}

	b, _ := json.Marshal(items)
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

// parseGWParam parses a comma-separated or repeated gw query param.
// e.g. ?gw=1,2,3 or ?gw=1&gw=2&gw=3
func parseGWParam(raw string) []int {
	if raw == "" {
		return nil
	}
	var out []int
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if v, err := strconv.Atoi(part); err == nil && v > 0 {
			out = append(out, v)
		}
	}
	return out
}
