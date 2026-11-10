package shotzone

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store persists open-play shot events and the two materialised boards. It
// owns only the sz_* tables and never touches players/teams/fixtures.
type Store struct {
	db *pgxpool.Pool
	v2 bool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// NewStoreV2 targets the FotMob-sourced sz2_* tables (migration 014). The SQL is
// shared with v1; sql() swaps the table names so both stay in lockstep.
func NewStoreV2(db *pgxpool.Pool) *Store {
	return &Store{db: db, v2: true}
}

// Longest names first: "sz_taken_board" must not be rewritten by a shorter key.
var v2Tables = strings.NewReplacer(
	"sz_events", "sz2_events",
	"sz_taken_board", "sz2_taken_board",
	"sz_conceded_board", "sz2_conceded_board",
)

func (s *Store) sql(q string) string {
	if s.v2 {
		return v2Tables.Replace(q)
	}
	return q
}

// UpsertEvents idempotently inserts qualifying open-play shots. The unique
// key (match_id, player_id, minute, zone) makes re-syncing a match a no-op.
func (s *Store) UpsertEvents(ctx context.Context, events []Event) error {
	if s.v2 {
		return s.upsertEventsV2(ctx, events)
	}
	for _, e := range events {
		_, err := s.db.Exec(ctx, s.sql(`
			INSERT INTO sz_events
				(match_id, season, match_date, minute, team_for, team_against, zone, player_id, player_name, is_goal, xg)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
			ON CONFLICT (match_id, player_id, minute, zone) DO UPDATE SET
				team_for     = EXCLUDED.team_for,
				team_against = EXCLUDED.team_against,
				player_name  = EXCLUDED.player_name,
				is_goal      = EXCLUDED.is_goal,
				xg           = EXCLUDED.xg
		`), e.MatchID, e.Season, e.MatchDate, e.Minute, e.TeamFor, e.TeamAgainst,
			string(e.Zone), e.PlayerID, e.PlayerName, e.IsGoal, e.XG)
		if err != nil {
			return fmt.Errorf("upsert sz_event match=%s player=%s: %w", e.MatchID, e.PlayerID, err)
		}
	}
	return nil
}

// upsertEventsV2 keys on FotMob's per-shot id: unlike v1's
// (match, player, minute, zone), two same-minute shots in one zone stay distinct.
func (s *Store) upsertEventsV2(ctx context.Context, events []Event) error {
	for _, e := range events {
		_, err := s.db.Exec(ctx, `
			INSERT INTO sz2_events
				(match_id, shot_id, season, match_date, minute, team_for, team_against, zone, player_id, player_name, is_goal, xg)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
			ON CONFLICT (match_id, shot_id) DO UPDATE SET
				team_for     = EXCLUDED.team_for,
				team_against = EXCLUDED.team_against,
				zone         = EXCLUDED.zone,
				player_name  = EXCLUDED.player_name,
				is_goal      = EXCLUDED.is_goal,
				xg           = EXCLUDED.xg
		`, e.MatchID, e.ShotID, e.Season, e.MatchDate, e.Minute, e.TeamFor, e.TeamAgainst,
			string(e.Zone), e.PlayerID, e.PlayerName, e.IsGoal, e.XG)
		if err != nil {
			return fmt.Errorf("upsert sz2_event match=%s shot=%s: %w", e.MatchID, e.ShotID, err)
		}
	}
	return nil
}

// ExistingMatchIDs returns the set of match ids already parsed into sz_events
// for a season, so the syncer can skip re-fetching them.
func (s *Store) ExistingMatchIDs(ctx context.Context, season string) (map[string]bool, error) {
	rows, err := s.db.Query(ctx,
		s.sql(`SELECT DISTINCT match_id FROM sz_events WHERE season = $1`), season)
	if err != nil {
		return nil, fmt.Errorf("query existing match ids: %w", err)
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

// ReadEvents returns all events for a season.
func (s *Store) ReadEvents(ctx context.Context, season string) ([]Event, error) {
	rows, err := s.db.Query(ctx, s.sql(`
		SELECT match_id, season, match_date, minute, team_for, team_against, zone,
		       player_id, player_name, is_goal, xg
		FROM sz_events
		WHERE season = $1
	`), season)
	if err != nil {
		return nil, fmt.Errorf("read events: %w", err)
	}
	defer rows.Close()
	var out []Event
	for rows.Next() {
		var e Event
		var zone string
		if err := rows.Scan(&e.MatchID, &e.Season, &e.MatchDate, &e.Minute, &e.TeamFor, &e.TeamAgainst,
			&zone, &e.PlayerID, &e.PlayerName, &e.IsGoal, &e.XG); err != nil {
			return nil, err
		}
		e.Zone = Zone(zone)
		out = append(out, e)
	}
	return out, rows.Err()
}

// ReplaceTakenBoard atomically swaps sz_taken_board for the freshly computed
// rows. A recompute is cheap and always full-set, so delete-all + insert
// inside a transaction is simplest and keeps the board consistent for readers.
func (s *Store) ReplaceTakenBoard(ctx context.Context, rows []TakenRow) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, s.sql(`DELETE FROM sz_taken_board`)); err != nil {
		return fmt.Errorf("clear taken board: %w", err)
	}
	now := time.Now().UTC()
	for _, r := range rows {
		var lastSeen *time.Time
		if !r.LastSeen.IsZero() {
			lastSeen = &r.LastSeen
		}
		_, err := tx.Exec(ctx, s.sql(`
			INSERT INTO sz_taken_board
				(team, zone, player_id, player_name, rank, shots, goals, xg, last_seen, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		`), r.Team, string(r.Zone), r.PlayerID, r.PlayerName, r.Rank, r.Shots, r.Goals, r.XG, lastSeen, now)
		if err != nil {
			return fmt.Errorf("insert taken board row: %w", err)
		}
	}
	return tx.Commit(ctx)
}

// ReplaceConcededBoard atomically swaps sz_conceded_board for the freshly
// computed rows.
func (s *Store) ReplaceConcededBoard(ctx context.Context, rows []ConcededRow) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, s.sql(`DELETE FROM sz_conceded_board`)); err != nil {
		return fmt.Errorf("clear conceded board: %w", err)
	}
	now := time.Now().UTC()
	for _, r := range rows {
		var lastSeen *time.Time
		if !r.LastSeen.IsZero() {
			lastSeen = &r.LastSeen
		}
		_, err := tx.Exec(ctx, s.sql(`
			INSERT INTO sz_conceded_board
				(team, zone, shots, goals, xg, last_seen, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
		`), r.Team, string(r.Zone), r.Shots, r.Goals, r.XG, lastSeen, now)
		if err != nil {
			return fmt.Errorf("insert conceded board row: %w", err)
		}
	}
	return tx.Commit(ctx)
}

// ListTeams returns the distinct teams observed (as shooters), sorted, for
// the frontend's team selector.
func (s *Store) ListTeams(ctx context.Context) ([]string, error) {
	rows, err := s.db.Query(ctx, s.sql(`SELECT DISTINCT team FROM sz_taken_board ORDER BY team`))
	if err != nil {
		return nil, fmt.Errorf("list teams: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// BoardUpdatedAt returns when the boards were last recomputed. Zero time when
// empty.
func (s *Store) BoardUpdatedAt(ctx context.Context) (time.Time, error) {
	var t *time.Time
	if err := s.db.QueryRow(ctx, s.sql(`SELECT max(updated_at) FROM sz_taken_board`)).Scan(&t); err != nil {
		return time.Time{}, fmt.Errorf("board updated_at: %w", err)
	}
	if t == nil {
		return time.Time{}, nil
	}
	return *t, nil
}

// ReadTakenBoard returns every taken-board row for a team (all zones incl.
// overall).
func (s *Store) ReadTakenBoard(ctx context.Context, team string) ([]TakenRow, error) {
	rows, err := s.db.Query(ctx, s.sql(`
		SELECT team, zone, player_id, player_name, rank, shots, goals, xg, last_seen
		FROM sz_taken_board
		WHERE team = $1
		ORDER BY zone, rank
	`), team)
	if err != nil {
		return nil, fmt.Errorf("read taken board: %w", err)
	}
	defer rows.Close()
	return scanTaken(rows)
}

// ReadConcededBoard returns every conceded-board row for a team (all zones
// incl. overall).
func (s *Store) ReadConcededBoard(ctx context.Context, team string) ([]ConcededRow, error) {
	rows, err := s.db.Query(ctx, s.sql(`
		SELECT team, zone, shots, goals, xg, last_seen
		FROM sz_conceded_board
		WHERE team = $1
		ORDER BY zone
	`), team)
	if err != nil {
		return nil, fmt.Errorf("read conceded board: %w", err)
	}
	defer rows.Close()
	var out []ConcededRow
	for rows.Next() {
		var r ConcededRow
		var zone string
		var lastSeen *time.Time
		if err := rows.Scan(&r.Team, &zone, &r.Shots, &r.Goals, &r.XG, &lastSeen); err != nil {
			return nil, err
		}
		r.Zone = Zone(zone)
		if lastSeen != nil {
			r.LastSeen = *lastSeen
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func scanTaken(rows pgx.Rows) ([]TakenRow, error) {
	var out []TakenRow
	for rows.Next() {
		var r TakenRow
		var zone string
		var lastSeen *time.Time
		if err := rows.Scan(&r.Team, &zone, &r.PlayerID, &r.PlayerName, &r.Rank, &r.Shots, &r.Goals, &r.XG, &lastSeen); err != nil {
			return nil, err
		}
		r.Zone = Zone(zone)
		if lastSeen != nil {
			r.LastSeen = *lastSeen
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// TeamShotTotal is one team's overall open-play shots taken and conceded over
// its rolling window.
type TeamShotTotal struct {
	Team          string
	Taken         int
	TakenGoals    int
	TakenXG       float64
	Conceded      int
	ConcededGoals int
	ConcededXG    float64
}

// ReadTeamTotals returns every team's overall taken/conceded totals in one
// query, for the league-wide overview.
func (s *Store) ReadTeamTotals(ctx context.Context) ([]TeamShotTotal, error) {
	rows, err := s.db.Query(ctx, s.sql(`
		SELECT t.team, t.shots, t.goals, t.xg,
		       COALESCE(c.shots, 0), COALESCE(c.goals, 0), COALESCE(c.xg, 0)
		FROM (
			SELECT team, sum(shots)::int AS shots, sum(goals)::int AS goals, sum(xg)::float8 AS xg
			FROM sz_taken_board
			WHERE zone = 'overall'
			GROUP BY team
		) t
		LEFT JOIN sz_conceded_board c ON c.team = t.team AND c.zone = 'overall'
		ORDER BY t.team
	`))
	if err != nil {
		return nil, fmt.Errorf("read team totals: %w", err)
	}
	defer rows.Close()
	var out []TeamShotTotal
	for rows.Next() {
		var r TeamShotTotal
		if err := rows.Scan(&r.Team, &r.Taken, &r.TakenGoals, &r.TakenXG, &r.Conceded, &r.ConcededGoals, &r.ConcededXG); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
