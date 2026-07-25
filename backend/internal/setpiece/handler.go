package setpiece

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/go-chi/chi/v5"
)

// confFullShots is the shot/attempt count at which a signal is treated as fully
// confident. Confidence scales linearly up to this and caps at 1.0.
const confFullShots = 5.0

// cacheTTL for the assembled API responses.
const responseCacheTTL = 30 * time.Minute

// Handler serves the read API under /api/setpiece. It is wired independently of
// fantasy.Source (docs §3).
type Handler struct {
	store         *Store
	cache         Cache
	windowMatches int
}

func NewHandler(store *Store, cache Cache, windowMatches int) *Handler {
	return &Handler{store: store, cache: cache, windowMatches: windowMatches}
}

// --- response DTOs ---------------------------------------------------------

type takerRow struct {
	PlayerID   string     `json:"player_id"`
	PlayerName string     `json:"player_name"`
	Rank       int        `json:"rank"`
	Attempts   int        `json:"attempts"`
	Goals      int        `json:"goals"`
	LastTaken  *time.Time `json:"last_taken,omitempty"`
	Confidence float64    `json:"confidence"`
}

type targetRow struct {
	PlayerID      string     `json:"player_id"`
	PlayerName    string     `json:"player_name"`
	Rank          int        `json:"rank"`
	Shots         int        `json:"shots"`
	Goals         int        `json:"goals"`
	XG            float64    `json:"xg"`
	HeaderPct     *float64   `json:"header_pct,omitempty"`
	WeightedScore float64    `json:"weighted_score"`
	Duty          string     `json:"duty"`
	LastSeen      *time.Time `json:"last_seen,omitempty"`
	Confidence    float64    `json:"confidence"`
}

type takerSet struct {
	Penalty []takerRow `json:"penalty"`
	DFK     []takerRow `json:"dfk"`
}

type teamGroup struct {
	Team    string      `json:"team"`
	Takers  takerSet    `json:"takers"`
	Targets []targetRow `json:"targets"` // duty='all', ranked
	// TargetsByDuty is the per-duty breakdown ('corner'/'setpiece'), included in
	// the list payload so the frontend can drive its duty filter client-side
	// without a second fetch (docs §5).
	TargetsByDuty map[string][]targetRow `json:"targets_by_duty"`
}

type teamDetail struct {
	teamGroup
	RecentEvents []eventRow `json:"recent_events"`
}

// teamsResponse wraps the team groups with the window/freshness meta the
// frontend card needs ("last N matches · updated Xh ago").
type teamsResponse struct {
	WindowMatches int         `json:"window_matches"`
	UpdatedAt     *time.Time  `json:"updated_at,omitempty"`
	Teams         []teamGroup `json:"teams"`
}

type eventRow struct {
	MatchDate  time.Time `json:"match_date"`
	Minute     int       `json:"minute"`
	Role       string    `json:"role"`
	Duty       string    `json:"duty"`
	PlayerName string    `json:"player_name"`
	IsHeader   bool      `json:"is_header"`
	XG         float64   `json:"xg"`
}

// Teams handles GET /api/setpiece/teams.
func (h *Handler) Teams(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	const cacheKey = "setpiece:teams"

	if b := h.cacheGet(ctx, cacheKey); b != nil {
		writeRaw(w, b)
		return
	}

	rows, err := h.store.ReadBoard(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}

	resp := teamsResponse{
		WindowMatches: h.windowMatches,
		Teams:         groupTeams(rows),
	}
	if ts, err := h.store.BoardUpdatedAt(ctx); err == nil && !ts.IsZero() {
		resp.UpdatedAt = &ts
	}

	b, _ := json.Marshal(resp)
	h.cacheSet(ctx, cacheKey, b)
	writeRaw(w, b)
}

// Team handles GET /api/setpiece/teams/{understat_team}.
func (h *Handler) Team(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	team := chi.URLParam(r, "understat_team")

	board, err := h.store.ReadTeamBoard(ctx, team)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	if len(board) == 0 {
		respondError(w, http.StatusNotFound, "team not found")
		return
	}

	groups := groupTeams(board)
	detail := teamDetail{teamGroup: groups[0]}

	// Recent events for this team (best-effort; not fatal).
	if evs, err := h.recentTeamEvents(ctx, team); err == nil {
		detail.RecentEvents = evs
	}

	respondJSON(w, http.StatusOK, detail)
}

// recentTeamEvents returns the 20 most recent events for a team from the board's
// backing table.
func (h *Handler) recentTeamEvents(ctx context.Context, team string) ([]eventRow, error) {
	rows, err := h.store.db.Query(ctx, `
		SELECT match_date, minute, role, duty, player_name, is_header, xg
		FROM sp_events
		WHERE understat_team = $1
		ORDER BY match_date DESC, minute DESC
		LIMIT 20
	`, team)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []eventRow
	for rows.Next() {
		var e eventRow
		if err := rows.Scan(&e.MatchDate, &e.Minute, &e.Role, &e.Duty, &e.PlayerName, &e.IsHeader, &e.XG); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// groupTeams folds board rows into per-team groups. Rows arrive pre-ranked.
func groupTeams(rows []BoardRow) []teamGroup {
	byTeam := map[string]*teamGroup{}
	order := []string{}
	for _, r := range rows {
		g := byTeam[r.UnderstatTeam]
		if g == nil {
			// Initialise slices so absent signals marshal as [] not null, which
			// the frontend relies on (e.g. takers.penalty.length).
			g = &teamGroup{
				Team:          r.UnderstatTeam,
				Takers:        takerSet{Penalty: []takerRow{}, DFK: []takerRow{}},
				Targets:       []targetRow{},
				TargetsByDuty: map[string][]targetRow{},
			}
			byTeam[r.UnderstatTeam] = g
			order = append(order, r.UnderstatTeam)
		}
		switch {
		case r.Role == RoleTaker && r.Duty == DutyPenalty:
			g.Takers.Penalty = append(g.Takers.Penalty, toTakerRow(r))
		case r.Role == RoleTaker && r.Duty == DutyDFK:
			g.Takers.DFK = append(g.Takers.DFK, toTakerRow(r))
		case r.Role == RoleTarget && r.Duty == DutyAll:
			g.Targets = append(g.Targets, toTargetRow(r))
		case r.Role == RoleTarget:
			g.TargetsByDuty[string(r.Duty)] = append(g.TargetsByDuty[string(r.Duty)], toTargetRow(r))
		}
	}
	sort.Strings(order)
	out := make([]teamGroup, 0, len(order))
	for _, t := range order {
		out = append(out, *byTeam[t])
	}
	return out
}

func toTakerRow(r BoardRow) takerRow {
	tr := takerRow{
		PlayerID:   r.PlayerID,
		PlayerName: r.PlayerName,
		Rank:       r.Rank,
		Attempts:   r.RawCount,
		Goals:      r.Goals,
		Confidence: confidence(r.RawCount),
	}
	if !r.LastSeen.IsZero() {
		ls := r.LastSeen
		tr.LastTaken = &ls
	}
	return tr
}

func toTargetRow(r BoardRow) targetRow {
	tr := targetRow{
		PlayerID:      r.PlayerID,
		PlayerName:    r.PlayerName,
		Rank:          r.Rank,
		Shots:         r.RawCount,
		Goals:         r.Goals,
		XG:            r.XGSum,
		HeaderPct:     r.HeaderPct,
		WeightedScore: r.WeightedScore,
		Duty:          string(r.Duty),
		Confidence:    confidence(r.RawCount),
	}
	if !r.LastSeen.IsZero() {
		ls := r.LastSeen
		tr.LastSeen = &ls
	}
	return tr
}

// confidence scales linearly with observed volume and caps at 1.0.
func confidence(rawCount int) float64 {
	c := float64(rawCount) / confFullShots
	if c > 1 {
		c = 1
	}
	return round2(c)
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
