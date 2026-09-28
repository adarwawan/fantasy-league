package shotzone

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
)

// responseCacheTTL for the assembled API responses.
const responseCacheTTL = 30 * time.Minute

// Handler serves the read API under /api/shot-zones. Isolated namespace, like
// setpiece — never wired through fantasy.Source.
type Handler struct {
	store         *Store
	cache         Cache
	windowMatches int
}

func NewHandler(store *Store, cache Cache, windowMatches int) *Handler {
	return &Handler{store: store, cache: cache, windowMatches: windowMatches}
}

// --- response DTOs ---------------------------------------------------------

type takenRow struct {
	PlayerID   string     `json:"player_id"`
	PlayerName string     `json:"player_name"`
	Rank       int        `json:"rank"`
	Shots      int        `json:"shots"`
	Goals      int        `json:"goals"`
	XG         float64    `json:"xg"`
	LastSeen   *time.Time `json:"last_seen,omitempty"`
}

type concededTotals struct {
	Shots    int        `json:"shots"`
	Goals    int        `json:"goals"`
	XG       float64    `json:"xg"`
	LastSeen *time.Time `json:"last_seen,omitempty"`
}

// zoneTakenSet groups a team's own shooters, one list per zone plus the
// cross-zone overall list — sent together so the frontend can switch its
// zone tabs client-side without a second fetch (same pattern as setpiece's
// targets_by_duty).
type zoneTakenSet struct {
	Overall []takenRow `json:"overall"`
	Center  []takenRow `json:"center"`
	Left    []takenRow `json:"left"`
	Right   []takenRow `json:"right"`
	Outside []takenRow `json:"outside"`
}

type zoneConcededSet struct {
	Overall concededTotals `json:"overall"`
	Center  concededTotals `json:"center"`
	Left    concededTotals `json:"left"`
	Right   concededTotals `json:"right"`
	Outside concededTotals `json:"outside"`
}

type teamsResponse struct {
	WindowMatches int        `json:"window_matches"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
	Teams         []string   `json:"teams"`
}

type teamDetail struct {
	Team          string          `json:"team"`
	WindowMatches int             `json:"window_matches"`
	UpdatedAt     *time.Time      `json:"updated_at,omitempty"`
	Taken         zoneTakenSet    `json:"taken"`
	Conceded      zoneConcededSet `json:"conceded"`
}

// Teams handles GET /api/shot-zones/teams.
func (h *Handler) Teams(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	const cacheKey = "shotzone:teams"

	if b := h.cacheGet(ctx, cacheKey); b != nil {
		writeRaw(w, b)
		return
	}

	teams, err := h.store.ListTeams(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}

	resp := teamsResponse{WindowMatches: h.windowMatches, Teams: teams}
	if ts, err := h.store.BoardUpdatedAt(ctx); err == nil && !ts.IsZero() {
		resp.UpdatedAt = &ts
	}

	b, _ := json.Marshal(resp)
	h.cacheSet(ctx, cacheKey, b)
	writeRaw(w, b)
}

// Team handles GET /api/shot-zones/teams/{team}.
func (h *Handler) Team(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	team := teamParam(r)

	taken, err := h.store.ReadTakenBoard(ctx, team)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	conceded, err := h.store.ReadConcededBoard(ctx, team)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	if len(taken) == 0 && len(conceded) == 0 {
		respondError(w, http.StatusNotFound, "team not found")
		return
	}

	detail := teamDetail{
		Team:          team,
		WindowMatches: h.windowMatches,
		Taken:         groupTaken(taken),
		Conceded:      groupConceded(conceded),
	}
	if ts, err := h.store.BoardUpdatedAt(ctx); err == nil && !ts.IsZero() {
		detail.UpdatedAt = &ts
	}

	respondJSON(w, http.StatusOK, detail)
}

func groupTaken(rows []TakenRow) zoneTakenSet {
	// Initialise slices so an absent zone marshals as [] not null, which the
	// frontend relies on (e.g. `.length`).
	set := zoneTakenSet{
		Overall: []takenRow{}, Center: []takenRow{}, Left: []takenRow{},
		Right: []takenRow{}, Outside: []takenRow{},
	}
	for _, r := range rows {
		row := toTakenRow(r)
		switch r.Zone {
		case ZoneOverall:
			set.Overall = append(set.Overall, row)
		case ZoneCenter:
			set.Center = append(set.Center, row)
		case ZoneLeft:
			set.Left = append(set.Left, row)
		case ZoneRight:
			set.Right = append(set.Right, row)
		case ZoneOutside:
			set.Outside = append(set.Outside, row)
		}
	}
	return set
}

func groupConceded(rows []ConcededRow) zoneConcededSet {
	set := zoneConcededSet{}
	for _, r := range rows {
		totals := toConcededTotals(r)
		switch r.Zone {
		case ZoneOverall:
			set.Overall = totals
		case ZoneCenter:
			set.Center = totals
		case ZoneLeft:
			set.Left = totals
		case ZoneRight:
			set.Right = totals
		case ZoneOutside:
			set.Outside = totals
		}
	}
	return set
}

func toTakenRow(r TakenRow) takenRow {
	tr := takenRow{
		PlayerID:   r.PlayerID,
		PlayerName: r.PlayerName,
		Rank:       r.Rank,
		Shots:      r.Shots,
		Goals:      r.Goals,
		XG:         r.XG,
	}
	if !r.LastSeen.IsZero() {
		ls := r.LastSeen
		tr.LastSeen = &ls
	}
	return tr
}

func toConcededTotals(r ConcededRow) concededTotals {
	ct := concededTotals{Shots: r.Shots, Goals: r.Goals, XG: r.XG}
	if !r.LastSeen.IsZero() {
		ls := r.LastSeen
		ct.LastSeen = &ls
	}
	return ct
}

// teamParam reads the {team} path segment, decoding it defensively: when a
// segment mixes an unescaped character (e.g. the apostrophe in "Nott'm
// Forest") with a percent-escaped one (the space), Go's net/url sets
// URL.RawPath and chi's URLParam returns that raw, still-escaped segment
// instead of the decoded one. PathUnescape is a no-op on an already-decoded
// value, so this is safe for the common case too.
func teamParam(r *http.Request) string {
	team := chi.URLParam(r, "team")
	if decoded, err := url.PathUnescape(team); err == nil {
		return decoded
	}
	return team
}

func (h *Handler) cacheGet(ctx context.Context, key string) []byte {
	if h.cache == nil {
		return nil
	}
	b, err := h.cache.Get(ctx, key)
	if err != nil {
		return nil
	}
	return b
}

func (h *Handler) cacheSet(ctx context.Context, key string, b []byte) {
	if h.cache != nil {
		_ = h.cache.Set(ctx, key, b, responseCacheTTL)
	}
}

func writeRaw(w http.ResponseWriter, b []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}
