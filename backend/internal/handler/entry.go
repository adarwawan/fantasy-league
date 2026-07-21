package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"fantasy-league/internal/fantasy"
)

type entryPlayerStore interface {
	PlayerIDsByExternalIDs(ctx context.Context, gameID string, externalIDs []string) (map[string]string, error)
}

// EntryHandler handles GET /api/{game}/entry/{id}, loading a manager's current
// squad as planner seed data. Support is per-game: a game is loadable iff its
// source implements fantasy.EntryLoader, so new games need no changes here.
type EntryHandler struct {
	loaders map[string]fantasy.EntryLoader
	store   entryPlayerStore
}

// NewEntryHandler registers every source that supports team loading. Callers
// pass all sources; those implementing fantasy.EntryLoader are enabled.
func NewEntryHandler(sources []fantasy.Source, store entryPlayerStore) *EntryHandler {
	loaders := make(map[string]fantasy.EntryLoader)
	for _, src := range sources {
		if el, ok := src.(fantasy.EntryLoader); ok {
			loaders[src.GameID()] = el
		}
	}
	return &EntryHandler{loaders: loaders, store: store}
}

// SupportedGames returns the game IDs that support team loading, so the API/UI
// can advertise the capability without trial-and-error.
func (h *EntryHandler) SupportedGames() []string {
	games := make([]string, 0, len(h.loaders))
	for g := range h.loaders {
		games = append(games, g)
	}
	return games
}

type entryPickJSON struct {
	PlayerID   string `json:"player_id"`
	IsCaptain  bool   `json:"is_captain"`
	Multiplier int    `json:"multiplier"`
}

type entryResponse struct {
	EntryID   string          `json:"entry_id"`
	TeamValue float64         `json:"team_value"`
	Bank      float64         `json:"bank"`
	GW        int             `json:"gw"`
	Picks     []entryPickJSON `json:"picks"`
}

func (h *EntryHandler) Load(w http.ResponseWriter, r *http.Request) {
	game := chi.URLParam(r, "game")
	source, ok := h.loaders[game]
	if !ok {
		respondError(w, http.StatusNotFound, "loading a team is not supported for this game")
		return
	}

	entryID := chi.URLParam(r, "id")
	if _, err := strconv.Atoi(entryID); err != nil {
		respondError(w, http.StatusBadRequest, "entry id must be a number")
		return
	}

	ctx := r.Context()
	gw, err := source.CurrentGW(ctx)
	if err != nil || gw == 0 {
		respondError(w, http.StatusBadGateway, "could not determine current gameweek")
		return
	}

	picks, err := source.FetchPicks(ctx, entryID, gw)
	if err != nil {
		respondError(w, http.StatusBadGateway, "could not load this team — check the FPL ID")
		return
	}
	if len(picks) == 0 {
		respondError(w, http.StatusNotFound, "no picks found for this team")
		return
	}

	extIDs := make([]string, len(picks))
	for i, p := range picks {
		extIDs[i] = p.PlayerID
	}
	idMap, err := h.store.PlayerIDsByExternalIDs(ctx, game, extIDs)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not resolve picked players")
		return
	}

	out := make([]entryPickJSON, 0, len(picks))
	for _, p := range picks {
		id, ok := idMap[p.PlayerID]
		if !ok {
			continue // player not in our DB (e.g. just transferred out of the game)
		}
		out = append(out, entryPickJSON{PlayerID: id, IsCaptain: p.IsCaptain, Multiplier: p.Multiplier})
	}

	summary, err := source.FetchEntrySummary(ctx, entryID)
	if err != nil {
		respondError(w, http.StatusBadGateway, "could not load this team's budget")
		return
	}

	respondJSON(w, http.StatusOK, entryResponse{
		EntryID:   entryID,
		TeamValue: summary.TeamValue,
		Bank:      summary.Bank,
		GW:        gw,
		Picks:     out,
	})
}
